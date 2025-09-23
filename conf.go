package main

import (
	"errors"
	"net"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Pool conf object
type PoolConf struct {
	Name string `yaml:"name"`
	MyIp string `yaml:"myip"`

	Network string `yaml:"network"`
	Subnet  string `yaml:"subnet"`
	Netmask string `yaml:"mask"`

	Start string `yaml:"start"`
	End   string `yaml:"end"`

	Router []string `yaml:"routers"`
	Dns    []string `yaml:"dns"`

	LeaseTime uint32 `yaml:"leasetime"`

	// TODO: add arbitrary options aside from just router/dns

	ReservedHosts []HostConf `yaml:"hosts"`
}

func (pc PoolConf) ToPool() (*Pool, error) {
	pool := NewPool()

	pool.Name = pc.Name

	if strings.Contains(pool.Name, "/") {
		return nil, errors.New("Pool names cannot contain slashes as they are used in file names")
	}

	// FIXME: input validation for all of these
	pool.Network = net.ParseIP(pc.Network)
	pool.Netmask = net.ParseIP(pc.Netmask)
	pool.Start = net.ParseIP(pc.Start)
	pool.End = net.ParseIP(pc.End)
	pool.MyIp = IpToFixedV4(net.ParseIP(pc.MyIp))
	pool.LeaseTime = time.Second * time.Duration(pc.LeaseTime)

	pool.Broadcast = calcBroadcast(pool.Network, pool.Netmask)

	for _, ip := range pc.Router {
		pool.Router = append(pool.Router, net.ParseIP(ip))
	}

	for _, ip := range pc.Dns {
		pool.Dns = append(pool.Dns, net.ParseIP(ip))
	}

	for _, host := range pc.ReservedHosts {
		if err := pool.AddReservedHost(host.ToHost()); err != nil {
			return nil, err
		}
	}

	return pool, nil
}

type HostConf struct {
	IP       string `yaml:"ip"`
	Mac      string `yaml:"hw"`
	Hostname string `yaml:"hostname"`
	// TODO: add custom options scoped to host
}

func (hc *HostConf) ToHost() *ReservedHost {
	return &ReservedHost{
		Mac: StrToMac(hc.Mac),
		IP:  IpToFixedV4(net.ParseIP(hc.IP)),
	}
}

// Root yaml conf
type Conf struct {
	Pools                 []PoolConf `yaml:"pools"`
	Leasedir              string     `yaml:"leasedir"`
	Interfaces            []string   `yaml:"interfaces"`
	MaxConcurrentRequests int        `yaml:"max_concurrent_requests"`
	RequestTimeoutSeconds int        `yaml:"request_timeout_seconds"`
}

func ParseConf(path string) (*Conf, error) {
	conf := &Conf{}
	var err error
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(content, &conf); err != nil {
		return nil, err
	}

	// Set defaults if not specified
	if conf.MaxConcurrentRequests == 0 {
		conf.MaxConcurrentRequests = 50
	}
	if conf.RequestTimeoutSeconds == 0 {
		conf.RequestTimeoutSeconds = 5
	}

	return conf, nil
}
