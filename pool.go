package main

import (
	"errors"
	"fmt"
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

type ReservedHost struct {
	Mac      MacAddress
	Hostname string
	IP       FixedV4
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

	// Internal lease database
	leasesByMac map[MacAddress]*Lease
	leaseByIp   map[FixedV4]*Lease

	// Internal database of fixed mac addresses to IPs for hosts,
	// sourced from configuration
	reservedByMac map[MacAddress]*ReservedHost
	reservedByIp  map[FixedV4]*ReservedHost

	m sync.RWMutex
}

func NewPool() *Pool {
	p := &Pool{}
	p.clearLeases()
	p.clearReservedHosts()
	return p
}

// Hacky, terrible, naive impl. I want an ordered int set!
func (p *Pool) getFreeIp(mac MacAddress) (FixedV4, error) {

	// If there is a reserved IP for this mac address, use that
	if host, ok := p.reservedByMac[mac]; ok {
		return host.IP, nil
	}

	// Try to find the next free IP within our range, while keeping
	// track of the first expired lease we found, in case we have no
	// otherwise free IPs
	start := IpToFixedV4(p.Start)
	end := IpToFixedV4(p.End)

	var foundExpired *Lease = nil

	for ipLong := start; ipLong <= end; ipLong++ {
		// Skip over any IPs in our range which are reserved
		if _, ok := p.reservedByIp[ipLong]; ok {
			continue
		}
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

func (p *Pool) clearReservedHosts() {
	p.reservedByMac = map[MacAddress]*ReservedHost{}
	p.reservedByIp = map[FixedV4]*ReservedHost{}
}

func (p *Pool) insertReservedHost(host *ReservedHost) {
	p.reservedByMac[host.Mac] = host
	p.reservedByIp[host.IP] = host
}

func (p *Pool) AddReservedHost(host *ReservedHost) error {
	if _, ok := p.reservedByIp[host.IP]; ok {
		return fmt.Errorf("Reserved hosts with duplicate IP: %v", host.IP)
	}
	if _, ok := p.reservedByMac[host.Mac]; ok {
		return fmt.Errorf("Reserved hosts with duplicate MAC: %v", host.Mac)
	}
	p.insertReservedHost(host)
	return nil
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

	ip, err := p.getFreeIp(mac)
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
