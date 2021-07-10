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
	Name        string
	Network     net.IP
	Netmask     net.IP
	Broadcast   net.IP
	Start       net.IP
	End         net.IP
	MyIp        FixedV4
	Router      []net.IP
	Dns         []net.IP
	LeaseTime   time.Duration
	Persistence Persistence

	leasesByMac map[MacAddress]*Lease
	leaseByIp   map[FixedV4]*Lease
	m           sync.RWMutex
}

func NewPool() *Pool {
	p := &Pool{}
	p.clearLeases()
	return p
}

// Hacky, terrible, naive impl. I want an ordered int set!
func (p *Pool) getFreeIp() (FixedV4, error) {

	// Try to find the next free IP within our range, while keeping
	// track of the first expired lease we found, in case we have no
	// otherwise free IPs
	start := IpToFixedV4(p.Start)
	end := IpToFixedV4(p.End)

	var foundExpired *Lease = nil

	for ipLong := start; ipLong <= end; ipLong++ {
		if lease, ok := p.leaseByIp[ipLong]; !ok {
			return ipLong, nil
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

	return 0, ErrNoIps
}

func (p *Pool) clearLeases() {
	p.leasesByMac = map[MacAddress]*Lease{}
	p.leaseByIp = map[FixedV4]*Lease{}
}

func (p *Pool) insertLease(lease *Lease) {
	p.leasesByMac[lease.Mac] = lease
	p.leaseByIp[lease.IP] = lease
}

func (p *Pool) deleteLease(lease *Lease) {
	delete(p.leasesByMac, lease.Mac)
	delete(p.leaseByIp, lease.IP)
}

func (p *Pool) TouchLeaseByMac(mac MacAddress) (*Lease, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	if lease, ok := p.leasesByMac[mac]; ok {
		lease.BumpExpiry(p.LeaseTime)
		p.persistLeases()
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
	p.persistLeases()
	return lease, nil
}

func (p *Pool) ReleaseLeaseByMac(mac MacAddress) (*Lease, bool) {
	p.m.Lock()
	defer p.m.Unlock()

	if lease, ok := p.leasesByMac[mac]; ok {
		p.deleteLease(lease)
		p.persistLeases()
		return lease, true
	}

	return nil, false
}

func (p *Pool) LoadLeases() (int, error) {
	if p.Persistence == nil {
		return 0, nil
	}
	p.m.Lock()
	defer p.m.Unlock()

	leases, err := p.Persistence.LoadLeases()
	if err != nil {
		return 0, err
	}

	p.clearLeases()

	for _, lease := range leases {
		p.insertLease(lease)
	}
	return len(leases), nil
}

func (p *Pool) persistLeases() error {
	if p.Persistence == nil {
		return nil
	}

	return p.Persistence.PersistLeases(p.leaseByIp)
}
