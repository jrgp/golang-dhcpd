//go:build linux

package main

import (
	"net"
	"syscall"

	"golang.org/x/net/ipv4"
)

// setupSocketOptions configures Linux-specific socket options for DHCP server
func setupSocketOptions(ln *net.UDPConn) error {
	// Boilerplate to get additional OOB data with each incoming packet, which
	// includes the ID of the incoming interface
	file, err := ln.File()
	if err != nil {
		return err
	}

	return syscall.SetsockoptInt(int(file.Fd()), syscall.IPPROTO_IP, syscall.IP_PKTINFO, 1)
}

// oObToInterface parses out-of-band data to extract interface information on Linux
func oObToInterface(oob []byte) (*net.Interface, error) {
	cm := &ipv4.ControlMessage{}

	if err := cm.Parse(oob); err != nil {
		return nil, err
	}

	iface, err := net.InterfaceByIndex(cm.IfIndex)

	if err != nil {
		return nil, err
	}

	return iface, nil
}
