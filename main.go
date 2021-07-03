package main

import (
	"log"
	"net"
	"syscall"
)

type App struct {
	Pool    *Pool
	Routers []string
	Dns     []string
	MyIp    FixedV4
}

func main() {
	var err error

	nic, err := net.InterfaceByName("eth1")
	if err != nil {
		log.Fatalf("Cannot get interface: %v", err)
	}

	app := &App{
		Pool: NewPool(
			net.ParseIP("172.17.0.100"),
			net.ParseIP("172.17.0.200"),
			net.ParseIP("255.255.255.0"),
			[]net.IP{net.ParseIP("172.17.0.1")},
			[]net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("1.1.1.2")},
			600,
		),
		MyIp: IpToFixedV4(net.ParseIP("172.17.0.2")),
	}

	app.Pool.Nic = nic

	addr := net.UDPAddr{
		Port: 67,
		IP:   net.ParseIP("0.0.0.0"),
	}

	ln, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Failed listening: %v", err)
	}

	// Boilerplate to get additional OOB data with each incoming packet, which
	// includes the ID of the incoming interface
	file, err := ln.File()
	if err != nil {
		log.Fatalf("Failed getting socket descriptor: %v", err)
	}

	syscall.SetsockoptInt(int(file.Fd()), syscall.IPPROTO_IP, syscall.IP_PKTINFO, 1)

	ln.SetReadBuffer(1048576)

	buf := make([]byte, 1024)
	oob := make([]byte, 1024)

	for {
		len, ooblen, _, remote, err := ln.ReadMsgUDP(buf, oob)
		if err != nil {
			log.Printf("Failed accepting: %v", err)
			continue
		}
		go NewConnectionHandler(buf[:len], oob[:ooblen], remote, app).Handle()
	}
}
