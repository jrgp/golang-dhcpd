//
// Helpers for parsing and encoding a unified DHCP message,
// including the header and the options
//
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

type DHCPMessage struct {
	Header  *MessageHeader
	Options *Options
}

func NewDhcpMessage() *DHCPMessage {
	return &DHCPMessage{
		Options: NewOptions(),
		Header:  &MessageHeader{},
	}
}

func (m *DHCPMessage) Encode(buf *bytes.Buffer) error {
	err := binary.Write(buf, binary.LittleEndian, m.Header)
	if err != nil {
		return fmt.Errorf("Writing dhcp header to our payload: %v", err)
	}

	err = m.Options.Encode(buf)
	if err != nil {
		return fmt.Errorf("Writing dhcp options to our payload: %v", err)
	}

	return nil
}

func ParseDhcpMessage(buf []byte) (*DHCPMessage, error) {
	header := &MessageHeader{}
	reader := bytes.NewReader(buf)

	// Parse DHCP header
	err := binary.Read(reader, binary.LittleEndian, header)
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

	// Parse arbitrary options
	options := ParseOptions(reader)

	// Confusingly, the Op type can be overridden using an option
	if option, ok := options.Get(OPTION_MESSAGE_TYPE); ok {
		if option.Header.Length == 1 {
			header.Op = option.Data[0]
		}
	}

	// Similarly, so can the ClientAddr
	if option, ok := options.Get(OPTION_REQUESTED_IP); ok {
		if option.Header.Length == 4 {
			ip, err := BytesToFixedV4(option.Data)
			if err == nil {
				header.ClientAddr = ip
			} else {
				log.Printf("Failed converting byte stream to fixed v4")
			}
		}
	}

	return &DHCPMessage{
		Options: options,
		Header:  header,
	}, nil
}
