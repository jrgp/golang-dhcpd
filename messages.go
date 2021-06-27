package main

import (
	"fmt"
)

const (
	DHCPDISCOVER = iota + 1
	DHCPOFFER
	DHCPREQUEST
	DHCPACK
	DHCPNAK
	DHCPDECLINE
	DHCPRELEASE
	DHCPINFORM
	// And others
)

var opNames = map[byte]string{
	DHCPOFFER: "DHCPOFFER",
	DHCPACK:   "DHCPACK",
}

type MacAddress [6]byte

func (m MacAddress) String() string {
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x", m[0], m[1], m[2], m[3], m[4], m[5])
}

var Magic = [4]byte{99, 130, 83, 99}

type MessageHeader struct {
	Op          byte
	HType       byte
	HLen        byte
	Hops        byte
	Identifier  uint32
	Secs        uint16
	Flags       uint16
	ClientAddr  uint32
	YourAddr    uint32
	ServerAddr  uint32
	GatewayAddr uint32
	Mac         MacAddress
	MacPadding  [10]byte
	Hostname    [64]byte
	Filename    [128]byte
	Magic       [4]byte // FIXME: convert these 4 bytes to an int
}
