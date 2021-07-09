//
// Helpers for parsing the DHCP header payload
//
package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

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

//
// Header of a DHCP payload
//

var Magic uint32 = 0x63825363

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
	Magic       uint32
}

func (h *MessageHeader) Encode(buf *bytes.Buffer) error {
	// Set constant boilerplate
	h.Magic = Magic
	h.HType = 1
	h.HLen = 6
	return binary.Write(buf, binary.BigEndian, h)
}

func ParseMessageHeader(reader *bytes.Reader) (*MessageHeader, error) {
	header := &MessageHeader{}
	err := binary.Read(reader, binary.BigEndian, header)
	if err != nil {
		return nil, fmt.Errorf("Failed unpacking header into struct: %v", err)
	}

	// Verify sanity
	if header.HType != 1 {
		return nil, fmt.Errorf("Only type 1 (ethernet) supported, not %v", header.HType)
	}
	if header.HLen != 6 {
		return nil, fmt.Errorf("Only 6 len mac addresses supported, not %v", header.HLen)
	}
	if header.Magic != Magic {
		return nil, fmt.Errorf("Incorrect option magic")
	}

	return header, nil
}
