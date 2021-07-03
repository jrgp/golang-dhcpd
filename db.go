package main

import (
	"errors"
	"net"
	"sync"
)

type Lease struct {
	Mac        MacAddress
	Hostname   string
	IP         FixedV4
	Expiration float64
}

type Pool struct {
	Start     net.IP
	End       net.IP
	Mask      net.IP
	Router    []net.IP
	Dns       []net.IP
	LeaseTime uint32
	Nic       *net.Interface

	leasesByMac map[MacAddress]*Lease
	leaseByIp   map[FixedV4]*Lease
	m           sync.RWMutex
}

func NewPool(start, end, mask net.IP, router, dns []net.IP, leaseTime uint32) *Pool {
	return &Pool{
		Start:       start,
		End:         end,
		Mask:        mask,
		Router:      router,
		Dns:         dns,
		LeaseTime:   leaseTime,
		leasesByMac: map[MacAddress]*Lease{},
		leaseByIp:   map[FixedV4]*Lease{},
	}
}

// Hacky, terrible, naive impl. I want an ordered int set!
func (p *Pool) getNextIp() (FixedV4, error) {
	start := ip2long(p.Start.String())
	end := ip2long(p.End.String())
	var found FixedV4
	for ipInt := start; ipInt <= end; ipInt++ {
		found := IpToFixedV4(net.ParseIP(long2ip(ipInt)))
		if _, ok := p.leaseByIp[found]; !ok {
			return found, nil
		}
	}
	return found, errors.New("No free IPs")
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
