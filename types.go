package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

//
// net.IPNet suitable for a map key
//
type HashableIpNet struct {
	IP   FixedV4
	Mask FixedV4
}

func IpNet2HashableIpNet(inet *net.IPNet) (HashableIpNet, error) {
	var result HashableIpNet
	if len(inet.Mask) != net.IPv4len {
		return result, errors.New("Not v4")
	}
	result.IP = IpToFixedV4(inet.IP)
	result.Mask = FixedV4(binary.BigEndian.Uint32(inet.Mask))
	return result, nil
}

//
// Fixed-width big-endian integer to keep track of IPv4 IPs, as they appear over the wire
//
type FixedV4 uint32

func (v4 FixedV4) String() string {
	ip := long2ip(uint32(v4))
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func (v4 FixedV4) Bytes() []byte {
	return long2bytes(uint32(v4))
}

func (v4 FixedV4) Long() uint32 {
	return uint32(v4)
}

func (v4 FixedV4) NetIp() net.IP {
	return long2ip(uint32(v4))
}

func (v4 FixedV4) Empty() bool {
	return uint32(v4) == 0
}

func IpToFixedV4(ip net.IP) FixedV4 {
	b := ip.To4()
	return FixedV4(binary.BigEndian.Uint32(b[0:4]))
}

func BytesToFixedV4(b []byte) (FixedV4, error) {
	if len(b) != 4 {
		return 0, errors.New("Incorrect length")
	}
	return FixedV4(binary.BigEndian.Uint32(b[0:4])), nil
}

//
// Fixed-width byte array for mac addresses, as they appear over the wire
//
type MacAddress [6]byte

func (m MacAddress) String() string {
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x", m[0], m[1], m[2], m[3], m[4], m[5])
}

func StrToMac(str string) MacAddress {
	var m MacAddress

	parts := strings.Split(str, ":")
	if len(parts) != 6 {
		return m
	}

	parsePart := func(s string) byte {
		if n, err := strconv.ParseUint(s, 16, 8); err == nil {
			return byte(n)
		}
		return 0
	}

	m = MacAddress{
		parsePart(parts[0]),
		parsePart(parts[1]),
		parsePart(parts[2]),
		parsePart(parts[3]),
		parsePart(parts[4]),
		parsePart(parts[5]),
	}

	return m
}
