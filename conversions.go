package main

import (
	"bytes"
	"encoding/binary"
	"net"
)

// Quickly ripped from https://www.socketloop.com/tutorials/golang-convert-ip-address-string-to-long-unsigned-32-bit-integer
func ip2long(ip net.IP) uint32 {
	var long uint32
	binary.Read(bytes.NewBuffer(ip.To4()), binary.LittleEndian, &long)
	return long
}

func long2ip(ipIntLong uint32) net.IP {
	ipInt := int64(ipIntLong)
	b0 := byte(ipInt >> 24 & 0xff)
	b1 := byte(ipInt >> 16 & 0xff)
	b2 := byte(ipInt >> 8 & 0xff)
	b3 := byte(ipInt & 0xff)
	return net.IP{b3, b2, b1, b0}
}

// Quickly ripped from https://gist.github.com/chiro-hiro/2674626cebbcb5a676355b7aaac4972d
func long2bytes(ip uint32) []byte {
	r := make([]byte, 4)
	for i := uint32(0); i < 4; i++ {
		r[3-i] = byte((ip >> (8 * i)) & 0xff)
	}
	return r
}

func bytes2long(ip []byte) uint32 {
	var long uint32
	binary.Read(bytes.NewBuffer(ip), binary.LittleEndian, &long)
	return long
}

func calcBroadcast(network, netmask net.IP) net.IP {
	broadcast := ip2long(network) | ^ip2long(netmask)
	return long2ip(broadcast)
}
