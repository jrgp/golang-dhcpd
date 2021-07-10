package main

import (
	"golang.org/x/net/ipv4"

	"errors"
	"log"
	"net"
	"path/filepath"
)

type App struct {
	ipnet2pool map[HashableIpNet]*Pool
	interfaces map[string]struct{}
}

func NewApp() *App {
	return &App{
		ipnet2pool: map[HashableIpNet]*Pool{},
		interfaces: map[string]struct{}{},
	}
}

func (a *App) InitConf(conf *Conf) error {

	for _, iface := range conf.Interfaces {
		a.interfaces[iface] = struct{}{}
	}

	if len(a.interfaces) == 0 {
		return errors.New("No interfaces configured")
	}

	for _, pc := range conf.Pools {
		pool, err := pc.ToPool()
		if err != nil {
			return err
		}

		pool.Persistence = NewFilePersistence(filepath.Join(conf.Leasedir, pool.Name+".json"))

		count, err := pool.LoadLeases()
		if err != nil {
			return err
		}

		if count == 1 {
			log.Printf("Loaded pool %v with %v lease", pool.Name, count)
		} else {
			log.Printf("Loaded pool %v with %v leases", pool.Name, count)
		}

		err = a.insertPool(pool)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) insertPool(p *Pool) error {
	ipnet := HashableIpNet{
		IP:   IpToFixedV4(p.Network),
		Mask: IpToFixedV4(p.Netmask),
	}

	if _, ok := a.ipnet2pool[ipnet]; ok {
		return errors.New("Duplicate IP network between pools")
	}

	a.ipnet2pool[ipnet] = p

	return nil
}

func (a *App) oObToInterface(oob []byte) (*net.Interface, error) {
	cm := &ipv4.ControlMessage{}

	if err := cm.Parse(oob); err != nil {
		return nil, err
	}

	iface, err := net.InterfaceByIndex(cm.IfIndex)

	if err != nil {
		return nil, err
	}

	return iface, nil
}

func (a *App) findPoolByInterface(iface *net.Interface) (*Pool, error) {
	addrs, err := iface.Addrs()

	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		// FIXME: should we verify addr.Network() is first "ip+net" ?
		_, ipnet, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}

		hipnet, err := IpNet2HashableIpNet(ipnet)
		if err != nil {
			continue
		}

		if pool, ok := a.ipnet2pool[hipnet]; ok {
			return pool, nil
		}
	}

	return nil, errors.New("Not found")
}

func (a *App) DispatchMessage(myBuf, myOob []byte, remote *net.UDPAddr, localSocket *net.UDPConn) {
	iface, err := a.oObToInterface(myOob)
	if err != nil {
		log.Printf("Failed parsing interface out of OOB: %v", err)
		return
	}

	if _, ok := a.interfaces[iface.Name]; !ok {
		log.Printf("Ignoring DHCP traffic on unconfigured interface %v", iface.Name)
		return
	}

	pool, err := a.findPoolByInterface(iface)
	if err != nil {
		log.Printf("Can't find pool based on IPs bound to %v", iface.Name)
		return
	}

	if remote.Port != 68 {
		log.Printf("Ignoring DHCP packet with source port %d rather than 68", remote.Port)
		return
	}

	message, err := ParseDhcpMessage(myBuf)
	if err != nil {
		log.Printf("Failed parsing dhcp packet: %v", err)
		return
	}

	handler := NewRequestHandler(message, pool)

	response := handler.Handle()

	if response != nil {
		// FIXME: options to sending to unicast, sending to relay, etc. Move these send functions
		// somewhere else.
		handler.sendMessageBroadcast(response, localSocket)
	}
}
