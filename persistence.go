package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

type Persistence interface {
	LoadLeases() (map[FixedV4]*Lease, error)
	PersistLeases(map[FixedV4]*Lease) error
}

type FilePersistenceLease struct {
	Hostname   string
	IP         string
	Mac        string
	Expiration time.Time
}

type FilePersistence struct {
	path string
}

func NewFilePersistence(path string) *FilePersistence {
	return &FilePersistence{path}
}

// Load on-disk json leases into our in-memory format
func (p *FilePersistence) decode(orig map[string]*FilePersistenceLease) map[FixedV4]*Lease {
	result := map[FixedV4]*Lease{}
	for _, lease := range orig {
		result[IpToFixedV4(net.ParseIP(lease.IP))] = &Lease{
			Mac:        StrToMac(lease.Mac),
			Hostname:   lease.Hostname,
			IP:         IpToFixedV4(net.ParseIP(lease.IP)),
			Expiration: lease.Expiration,
		}
	}
	return result
}

// Encoded in-memory leases into our on-disk json format
func (p *FilePersistence) encode(leases map[FixedV4]*Lease) map[string]*FilePersistenceLease {
	result := map[string]*FilePersistenceLease{}
	for _, lease := range leases {
		result[lease.IP.String()] = &FilePersistenceLease{
			Mac:        lease.Mac.String(),
			Hostname:   lease.Hostname,
			IP:         lease.IP.String(),
			Expiration: lease.Expiration,
		}
	}
	return result
}

func (p *FilePersistence) LoadLeases() (map[FixedV4]*Lease, error) {
	leases := map[FixedV4]*Lease{}

	contents, err := ioutil.ReadFile(p.path)
	if err != nil {
		// If file doesn't exist, don't surface it as an error,
		// and instead just return an empty list
		if errors.Is(err, os.ErrNotExist) {
			return leases, nil
		}
		return nil, err
	}

	fromFile := map[string]*FilePersistenceLease{}

	err = json.Unmarshal(contents, &fromFile)

	if err != nil {
		return nil, err
	}

	return p.decode(fromFile), nil
}

func (p *FilePersistence) PersistLeases(leases map[FixedV4]*Lease) error {
	encoded := p.encode(leases)
	payload, err := json.MarshalIndent(encoded, "", "   ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(p.path, payload, 0644)
	if err != nil {
		log.Printf("Failed persisting leases to %v: %v", p.path, err)
	}
	return err
}
