package main

import (
	"errors"
	"log"
	"net"
	"path/filepath"
	"time"
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

// For non-relayed requests: find a pool by comparing nets to local nic
// IPs
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

// For relayed requests: find a pool by comparing giaddr to configured
// pool nets
func (a *App) findPoolbyGiaddr(giaddr FixedV4) (*Pool, error) {
	for _, pool := range a.ipnet2pool {
		ipnet := &net.IPNet{
			IP:   pool.Network,
			Mask: net.IPMask([]byte(pool.Netmask)),
		}
		if ipnet.Contains(giaddr.NetIp()) {
			return pool, nil
		}
	}

	return nil, errors.New("Not found")
}

func (a *App) DispatchMessageWithTimeout(timeout time.Duration, myBuf, myOob []byte, remote *net.UDPAddr, localSocket *net.UDPConn) {
	done := make(chan struct{})

	go func() {
		a.DispatchMessage(myBuf, myOob, remote, localSocket)
		close(done)
	}()

	select {
	case <-done:
		// Request completed normally
	case <-time.After(timeout):
		log.Printf("DHCP request timeout")
	}
}

func (a *App) DispatchMessage(myBuf, myOob []byte, remote *net.UDPAddr, localSocket *net.UDPConn) {
	// Sanity remote port check
	if remote.Port != 67 && remote.Port != 68 {
		log.Printf("Ignoring DHCP packet with source port %d rather than 67 or 68", remote.Port)
		return
	}

	var err error

	// Grab iface and verify we're configured to work on it
	iface, err := oObToInterface(myOob)
	if err != nil {
		log.Printf("Failed parsing interface out of OOB: %v", err)
		return
	}

	if _, ok := a.interfaces[iface.Name]; !ok {
		log.Printf("Ignoring DHCP traffic on unconfigured interface %v", iface.Name)
		return
	}

	// Parse entire dhcp message
	message, err := ParseDhcpMessage(myBuf)
	if err != nil {
		log.Printf("Failed parsing dhcp packet: %v", err)
		return
	}

	var pool *Pool

	// Relayed request. Find pool based on giaddr
	if !message.Header.GatewayAddr.Empty() {
		pool, err = a.findPoolbyGiaddr(message.Header.GatewayAddr)
		if err != nil {
			log.Printf("Can't find pool based on IPs bound to %v", iface.Name)
			return
		}

	} else {
		pool, err = a.findPoolByInterface(iface)
		if err != nil {
			log.Printf("Can't find pool based on IPs bound to %v", iface.Name)
			return
		}
	}

	handler := NewRequestHandler(message, pool)

	response := handler.Handle()

	if response != nil {
		// In the case of a relayed request, send the response unicast to the relaying server
		if !message.Header.GatewayAddr.Empty() {
			handler.sendMessageRelayed(response, message.Header.GatewayAddr, localSocket)
		} else {
			handler.sendMessageBroadcast(response, localSocket)
		}
	}
}
