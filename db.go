package main

import (
	"errors"
	"net"
	"sync"
)

type Lease struct {
	Mac        MacAddress
	Hostname   string
	IP         uint32
	Expiration float64
}

type Pool struct {
	Start     uint32
	End       uint32
	Mask      uint32
	Router    []uint32
	Dns       []uint32
	LeaseTime uint32
	Nic       *net.Interface

	leasesByMac map[MacAddress]*Lease
	leaseByIp   map[uint32]*Lease
	m           sync.RWMutex
}

func NewPool(start, end, mask uint32, router, dns []uint32, leaseTime uint32) *Pool {
	return &Pool{
		Start:       start,
		End:         end,
		Mask:        mask,
		Router:      router,
		Dns:         dns,
		LeaseTime:   leaseTime,
		leasesByMac: map[MacAddress]*Lease{},
		leaseByIp:   map[uint32]*Lease{},
	}
}

// Hacky, terrible, naive impl. I want an ordered int set!
func (p *Pool) getNextIp() (uint32, error) {
	for ip := p.Start; ip <= p.End; ip++ {
		if _, ok := p.leaseByIp[ip]; !ok {
			return ip, nil
		}
	}
	return 0, errors.New("No free IPs")
}

func (p *Pool) insertLease(lease *Lease) {
	p.leasesByMac[lease.Mac] = lease
	p.leaseByIp[lease.IP] = lease
}

func (p *Pool) GetLeaseByMac(mac MacAddress) (*Lease, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	if lease, ok := p.leasesByMac[mac]; ok {
		return lease, true
	}
	return nil, false
}

func (p *Pool) GetNextLease(mac MacAddress, hostname string) (*Lease, error) {
	p.m.Lock()
	defer p.m.Unlock()

	ip, err := p.getNextIp()
	if err != nil {
		return nil, err
	}
	lease := &Lease{
		IP:       ip,
		Hostname: hostname,
		Mac:      mac,
	}
	p.insertLease(lease)
	return lease, nil
}
