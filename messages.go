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

type MessageOption struct {
	Header struct {
		Code   byte
		Length byte
	}
	Data []byte
}

func (o *MessageOption) CalculateLength() error {
	length := len(o.Data)
	if length > 255 {
		return fmt.Errorf("Length of option %v value '%v' is too long", o.Header.Code, o.Data)
	}
	o.Header.Length = byte(length)
	return nil
}

func NewMessageOption(code byte, data []byte) (*MessageOption, error) {
	option := &MessageOption{
		Data: data,
	}
	option.Header.Code = code
	if err := option.CalculateLength(); err != nil {
		return nil, err
	}
	return option, nil
}
