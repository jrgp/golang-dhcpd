// Helpers for parsing the DHCP header payload
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

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
