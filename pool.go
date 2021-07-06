package main

import (
	"errors"
	"net"
	"sync"
	"time"
)

var ErrNoIps = errors.New("No free IPs")

type Lease struct {
	Mac        MacAddress
	Hostname   string
	IP         FixedV4
	Expiration time.Time
}

func (l *Lease) BumpExpiry(d time.Duration) {
	l.Expiration = time.Now().Add(d)
}

func (l *Lease) Expired() bool {
	return time.Now().After(l.Expiration)
}

type Pool struct {
	Name      string
	Network   net.IP
	Netmask   net.IP
	Broadcast net.IP
	Start     net.IP
	End       net.IP
	MyIp      FixedV4
	Router    []net.IP
	Dns       []net.IP
	LeaseTime time.Duration
	Interface string

	leasesByMac map[MacAddress]*Lease
	leaseByIp   map[uint32]*Lease
	m           sync.RWMutex
}

func NewPool() *Pool {
	return &Pool{
		leasesByMac: map[MacAddress]*Lease{},
		leaseByIp:   map[uint32]*Lease{},
	}
}

// Hacky, terrible, naive impl. I want an ordered int set!
func (p *Pool) getFreeIp() (FixedV4, error) {

	// Try to find the next free IP within our range, while keeping
	// track of the first expired lease we found, in case we have no
	// otherwise free IPs
	start := ip2long(p.Start)
	end := ip2long(p.End)

	var foundExpired *Lease = nil

	for ipLong := start; ipLong <= end; ipLong++ {
		if lease, ok := p.leaseByIp[ipLong]; !ok {
			return IpToFixedV4(long2ip(ipLong)), nil
		} else {
			if foundExpired == nil && lease.Expired() {
				foundExpired = lease
			}
		}
	}

	// We have a recovered expired lease. Delete it
	// and return its free IP
	if foundExpired != nil {
		p.deleteLease(foundExpired)
		return foundExpired.IP, nil
	}

	return FixedV4{}, ErrNoIps
}

func (p *Pool) insertLease(lease *Lease) {
	p.leasesByMac[lease.Mac] = lease
	p.leaseByIp[lease.IP.Long()] = lease
}

func (p *Pool) deleteLease(lease *Lease) {
	delete(p.leasesByMac, lease.Mac)
	delete(p.leaseByIp, lease.IP.Long())
}

func (p *Pool) TouchLeaseByMac(mac MacAddress) (*Lease, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	if lease, ok := p.leasesByMac[mac]; ok {
		lease.BumpExpiry(p.LeaseTime)
		return lease, true
	}
	return nil, false
}

func (p *Pool) GetNextLease(mac MacAddress, hostname string) (*Lease, error) {
	p.m.Lock()
	defer p.m.Unlock()

	ip, err := p.getFreeIp()
	if err != nil {
		return nil, err
	}
	lease := &Lease{
		IP:       ip,
		Hostname: hostname,
		Mac:      mac,
	}
	lease.BumpExpiry(p.LeaseTime)
	p.insertLease(lease)
	return lease, nil
}

func (p *Pool) ReleaseLeaseByMac(mac MacAddress) (*Lease, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	if lease, ok := p.leasesByMac[mac]; ok {
		p.deleteLease(lease)
		return lease, true
	}

	return nil, false
}
