package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"
)

// Quickly ripped from https://www.socketloop.com/tutorials/golang-convert-ip-address-string-to-long-unsigned-32-bit-integer
func ip2long(ip string) uint32 {
	var long uint32
	binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.LittleEndian, &long)
	return long
}

func long2ip(ipIntLong uint32) string {
	ipInt := int64(ipIntLong)
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}

// Quickly ripped from https://gist.github.com/chiro-hiro/2674626cebbcb5a676355b7aaac4972d
func long2bytes(ip uint32) []byte {
	r := make([]byte, 4)
	for i := uint32(0); i < 4; i++ {
		r[i] = byte((ip >> (8 * i)) & 0xff)
	}
	return r
}
