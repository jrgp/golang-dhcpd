//
// Helpers for parsing the DHCP header payload
//
package main

import (
	"errors"
	"fmt"
	"net"
)

//
// Fixed-width byte array to keep track of IPv4 IPs, as they appear over the wire
//
type FixedV4 [4]byte

func (v4 FixedV4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v4[0], v4[1], v4[2], v4[3])
}

func (v4 FixedV4) Bytes() []byte {
	return []byte{v4[0], v4[1], v4[2], v4[3]}
}

func IpToFixedV4(ip net.IP) FixedV4 {
	v4 := ip.To4()
	return FixedV4{v4[0], v4[1], v4[2], v4[3]}
}

func BytesToFixedV4(b []byte) (FixedV4, error) {
	if len(b) != 4 {
		return FixedV4{}, errors.New("Incorrect length")
	}
	return FixedV4{b[0], b[1], b[2], b[3]}, nil
}

//
// Fixed-width byte array for mac addresses, as they appear over the wire
//
type MacAddress [6]byte

func (m MacAddress) String() string {
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x", m[0], m[1], m[2], m[3], m[4], m[5])
}

var Magic = [4]byte{99, 130, 83, 99}

// Header of a DHCP payload

type MessageHeader struct {
	Op          byte
	HType       byte
	HLen        byte
	Hops        byte
	Identifier  uint32
	Secs        uint16
	Flags       uint16
	ClientAddr  FixedV4
	YourAddr    FixedV4
	ServerAddr  FixedV4
	GatewayAddr FixedV4
	Mac         MacAddress
	MacPadding  [10]byte
	Hostname    [64]byte
	Filename    [128]byte
	Magic       [4]byte // FIXME: convert these 4 bytes to an int
}
