package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

func getFlags() (string, bool) {
	conf := flag.String("conf", "", "Path to configuration yaml file")
	unprivileged := flag.Bool("unprivileged", false, "Run as unprivileged process with inherited socket")
	flag.Parse()
	return *conf, *unprivileged
}

func setupSocket(ln *net.UDPConn) {
	// Configure socket to receive additional OOB data with each incoming packet,
	// which includes the ID of the incoming interface
	file, err := ln.File()
	if err != nil {
		log.Fatalf("Failed getting socket descriptor: %v", err)
	}
	defer file.Close()

	syscall.SetsockoptInt(int(file.Fd()), syscall.IPPROTO_IP, syscall.IP_PKTINFO, 1)
	ln.SetReadBuffer(1048576)
}

func forkUnprivilegedWithSocket(ln *net.UDPConn, confPath, username, groupname string) error {
	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user lookup failed: %w", err)
	}

	g, err := user.LookupGroup(groupname)
	if err != nil {
		return fmt.Errorf("group lookup failed: %w", err)
	}

	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(g.Gid)

	file, err := ln.File()
	if err != nil {
		return fmt.Errorf("failed to get socket file: %w", err)
	}
	defer file.Close()

	cmd := exec.Command(os.Args[0], "-conf", confPath, "-unprivileged")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{file}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}

	log.Printf("Starting unprivileged process as %s:%s", username, groupname)
	return cmd.Run()
}

func main() {
	var err error

	confPath, unprivileged := getFlags()

	if confPath == "" {
		log.Fatalf("Configuration file path not given")
	}

	conf, err := ParseConf(confPath)
	if err != nil {
		log.Fatalf("Failed parsing conf: %v", err)
	}

	var ln *net.UDPConn

	// Check if we're the unprivileged child process with inherited socket
	if unprivileged {
		// Restore socket from file descriptor 3 (first ExtraFile)
		file := os.NewFile(3, "socket")
		conn, err := net.FileConn(file)
		if err != nil {
			log.Fatalf("Failed to restore socket from fd: %v", err)
		}
		ln = conn.(*net.UDPConn)
		file.Close() // Close the file, keep the connection
	} else {
		// We're the privileged parent process
		addr := net.UDPAddr{
			Port: 67,
			IP:   net.ParseIP("0.0.0.0"),
		}

		ln, err = net.ListenUDP("udp", &addr)
		if err != nil {
			log.Fatalf("Failed listening: %v", err)
		}

		// If we need to drop privileges, fork to unprivileged process
		if os.Getuid() == 0 && conf.RunAsUser != "" {
			err := forkUnprivilegedWithSocket(ln, confPath, conf.RunAsUser, conf.RunAsGroup)
			if err != nil {
				log.Fatalf("Failed to fork unprivileged: %v", err)
			}
			return // Parent process exits
		}
	}

	defer ln.Close()

	setupSocket(ln)

	app := NewApp()

	err = app.InitConf(conf)

	if err != nil {
		log.Fatalf("Failed initializing: %v", err)
	}

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
