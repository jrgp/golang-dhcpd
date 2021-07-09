package main

import (
	"golang.org/x/net/ipv4"

	"errors"
	"log"
	"net"
	"path/filepath"
)

type App struct {
	pools          []*Pool
	interface2Pool map[string]*Pool
}

func NewApp() *App {
	return &App{
		pools:          []*Pool{},
		interface2Pool: map[string]*Pool{},
	}
}

func (a *App) InitPools(conf *Conf) error {

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
			log.Printf("Loaded pool %v on interface %v with %v lease", pool.Name, pool.Interface, count)
		} else {
			log.Printf("Loaded pool %v on interface %v with %v leases", pool.Name, pool.Interface, count)
		}

		err = a.insertPool(pool)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) insertPool(p *Pool) error {
	if _, ok := a.interface2Pool[p.Interface]; ok {
		return errors.New("Interfaces may be used by only one pool")
	}

	a.interface2Pool[p.Interface] = p
	a.pools = append(a.pools, p)

	return nil
}

func (a *App) oObToInterface(oob []byte) (string, error) {
	cm := &ipv4.ControlMessage{}

	if err := cm.Parse(oob); err != nil {
		return "", err
	}

	Interface, err := net.InterfaceByIndex(cm.IfIndex)

	if err != nil {
		return "", err
	}

	return Interface.Name, nil
}

func (a *App) findPoolByInterface(Interface string) (*Pool, error) {
	pool, ok := a.interface2Pool[Interface]
	if !ok {
		return nil, errors.New("Unconfigured")
	}
	return pool, nil
}

func (a *App) DispatchMessage(myBuf, myOob []byte, remote *net.UDPAddr, localSocket *net.UDPConn) {
	Interface, err := a.oObToInterface(myOob)
	if err != nil {
		log.Printf("Failed parsing interface out of OOB: %v", err)
		return
	}

	pool, err := a.findPoolByInterface(Interface)
	if err != nil {
		log.Printf("Ignoring DHCP traffic on unconfigured interface %v", Interface)
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
