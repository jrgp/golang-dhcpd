package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"time"
)

type PoolConf struct {
	Name      string `yaml:"name"`
	Interface string `yaml:"interface"`
	MyIp      string `yaml:"myip"`

	Network string `yaml:"network"`
	Subnet  string `yaml:"subnet"`
	Netmask string `yaml:"mask"`

	Start string `yaml:"start"`
	End   string `yaml:"end"`

	Router []string `yaml:"routers"`
	Dns    []string `yaml:"dns"`

	LeaseTime uint32 `yaml:"leasetime"`
}

func (pc PoolConf) ToPool() (*Pool, error) {
	pool := NewPool()

	// FIXME: input validation for all of these
	pool.Name = pc.Name
	pool.Network = net.ParseIP(pc.Network)
	pool.Netmask = net.ParseIP(pc.Netmask)
	pool.Start = net.ParseIP(pc.Start)
	pool.End = net.ParseIP(pc.End)
	pool.MyIp = IpToFixedV4(net.ParseIP(pc.MyIp))
	pool.LeaseTime = time.Second * time.Duration(pc.LeaseTime)
	pool.Interface = pc.Interface

	pool.Broadcast = calcBroadcast(pool.Network, pool.Netmask)

	for _, ip := range pc.Router {
		pool.Router = append(pool.Router, net.ParseIP(ip))
	}

	for _, ip := range pc.Dns {
		pool.Dns = append(pool.Dns, net.ParseIP(ip))
	}

	log.Printf("Loaded pool %v on interface %v", pool.Name, pool.Interface)

	return pool, nil
}

type HostConf struct {
	Ip        string `yaml:"ip"`
	Mac       string `yaml:"hw"`
	MacParsed MacAddress
}

type Conf struct {
	Pools []PoolConf `yaml:"pools"`
}

func ParseConf(path string) (*Conf, error) {
	conf := &Conf{}
	var err error
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(content, &conf); err != nil {
		return nil, err
	}
	return conf, nil
}
