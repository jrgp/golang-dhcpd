//go:build darwin

package main

import (
	"net"
)

// setupSocketOptions is a stub for macOS - DHCP server functionality is limited
// This allows the application to build and run basic tests on macOS
func setupSocketOptions(ln *net.UDPConn) error {
	// On macOS, we can't set IP_PKTINFO, so we just return nil
	// This means the DHCP server won't work properly on macOS, but it will build
	return nil
}

// oObToInterface is a stub for macOS - returns a dummy interface for testing
// This allows the application to build and run basic tests on macOS
func oObToInterface(oob []byte) (*net.Interface, error) {
	// On macOS, we can't parse the OOB data properly, so we return a dummy interface
	// This means the DHCP server won't work properly on macOS, but it will build
	// For testing purposes, we'll return the first available interface
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	if len(interfaces) == 0 {
		return nil, err
	}

	return &interfaces[0], nil
}
