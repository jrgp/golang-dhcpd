package main

import (
	"log"
	"net"
	"syscall"
)

func main() {
	var err error

	app := NewApp()

	err = app.InitPools()

	if err != nil {
		log.Fatalf("Failed initializing pools: %v", err)
	}

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

		go app.DispatchMessage(buf[:len], oob[:ooblen], remote)
	}
}
