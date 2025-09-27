package main

import (
	"flag"
	"log"
	"net"
	"time"
)

func getConfPath() string {
	conf := flag.String("conf", "", "Path to configuration yaml file")
	flag.Parse()
	return *conf
}

func main() {
	var err error

	confPath := getConfPath()

	if confPath == "" {
		log.Fatalf("Configuration file path not given")
	}

	conf, err := ParseConf(confPath)
	if err != nil {
		log.Fatalf("Failed parsing conf: %v", err)
	}

	app := NewApp()

	err = app.InitConf(conf)

	if err != nil {
		log.Fatalf("Failed initializing: %v", err)
	}

	addr := net.UDPAddr{
		Port: 67,
		IP:   net.ParseIP("0.0.0.0"),
	}

	ln, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Failed listening: %v", err)
	}
	defer ln.Close()

	// Setup platform-specific socket options
	err = setupSocketOptions(ln)
	if err != nil {
		log.Fatalf("Failed setting up socket options: %v", err)
	}

	ln.SetReadBuffer(1048576)

	buf := make([]byte, 1024)
	oob := make([]byte, 1024)

	requestSem := make(chan struct{}, conf.MaxConcurrentRequests)
	for {
		len, ooblen, _, remote, err := ln.ReadMsgUDP(buf, oob)
		if err != nil {
			log.Printf("Failed accepting: %v", err)
			continue
		}

		select {
		case requestSem <- struct{}{}:
			go func() {
				defer func() { <-requestSem }()
				timeout := time.Duration(conf.RequestTimeoutSeconds) * time.Second
				app.DispatchMessageWithTimeout(timeout, buf[:len], oob[:ooblen], remote, ln)
			}()
		default:
			log.Printf("DHCP request dropped - server busy")
		}
	}
}
