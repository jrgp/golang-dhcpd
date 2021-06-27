package main

import (
	"log"
	"net"
)

type App struct {
	Pool    *Pool
	Routers []string
	Dns     []string
	MyIp    uint32
}

func main() {
	app := &App{
		Pool: NewPool(
			ip2long("172.17.0.100"),
			ip2long("172.17.0.200"),
			ip2long("255.255.255.0"),
			[]uint32{ip2long("172.17.0.1")},
			[]uint32{ip2long("1.1.1.1"), ip2long("1.1.1.2")},
			600,
		),
		MyIp: ip2long("172.17.0.2"),
	}

	addr := net.UDPAddr{
		Port: 67,
		IP:   net.ParseIP("0.0.0.0"),
	}

	ln, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Failed listening: %v", err)
	}

	ln.SetReadBuffer(1048576)

	for {
		buf := make([]byte, 1024)
		len, remote, err := ln.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Failed accepting: %v", err)
			continue
		}
		go NewConnectionHandler(buf[:len], remote, app).Handle()
	}
}
