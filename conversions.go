package main

import (
	"encoding/binary"
	"net"
)

// Quickly ripped from https://www.socketloop.com/tutorials/golang-convert-ip-address-string-to-long-unsigned-32-bit-integer
func ip2long(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4()[0:4])
}

func long2ip(data uint32) net.IP {
	b := long2bytes(data)
	return net.IP{b[0], b[1], b[2], b[3]}
}

func long2bytes(data uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, data)
	return b
}

func calcBroadcast(network, netmask net.IP) net.IP {
	broadcast := ip2long(network) | ^ip2long(netmask)
	return long2ip(broadcast)
}
