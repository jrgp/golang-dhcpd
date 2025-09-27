// Helpers for parsing the DHCP option payloads
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

// Represent a single DHCP option
type Option struct {
	Header struct {
		Code   byte
		Length byte
	}
	Data []byte
}

func (o *Option) CalculateLength() error {
	length := len(o.Data)
	if length > 255 {
		return fmt.Errorf("Length of option %v value '%v' is too long", o.Header.Code, o.Data)
	}
	o.Header.Length = byte(length)
	return nil
}

//
// Easily handle lists of options. Each can only appear once.
//

type Options struct {
	order []byte
	data  map[byte]Option
}

func NewOptions() *Options {
	return &Options{
		data:  map[byte]Option{},
		order: []byte{},
	}
}

func (o *Options) GetAll() map[byte]Option {
	return o.data
}

func (o *Options) Dump() {
	for _, key := range o.order {
		option := o.data[key]
		log.Printf("%v = %v (%+v)", key, optionNames[key], option.Data)
	}
}

// Abstract away boilerplate for common getting operations
func (o *Options) Get(code byte) (Option, bool) {
	option, ok := o.data[code]
	return option, ok
}

func (o *Options) GetByte(code byte) byte {
	if option, ok := o.data[code]; ok {
		if len(option.Data) == 1 {
			return option.Data[0]
		}
	}
	return 0
}

// Primarily used in testing
func (o *Options) GetFixedV4s(code byte) []FixedV4 {
	if option, ok := o.data[code]; ok {
		if len(option.Data)%4 == 0 {
			ips := make([]FixedV4, len(option.Data)/4)
			for i := 0; i < len(ips); i++ {
				if ip, err := BytesToFixedV4(option.Data[i*4 : i*4+4]); err == nil {
					ips[i] = ip
				}
			}
			return ips
		}
	}
	return nil
}

// Abstract away boilerplate for common IP setting operations
func (o *Options) SetIPs(code byte, ips ...net.IP) {
	if len(ips) == 0 {
		return
	}
	data := make([]byte, 0, 4*len(ips))
	for _, ip := range ips {
		data = append(data, IpToFixedV4(ip).Bytes()...)
	}
	o.Set(code, data)
}

func (o *Options) SetFixedV4s(code byte, ips ...FixedV4) {
	if len(ips) == 0 {
		return
	}
	data := make([]byte, 0, 4*len(ips))
	for _, ip := range ips {
		data = append(data, ip.Bytes()...)
	}
	o.Set(code, data)
}

// Set a single option
func (o *Options) Set(code byte, data []byte) {
	option := Option{
		Data: data,
	}
	option.Header.Code = code
	if err := option.CalculateLength(); err != nil {
		log.Printf("Can't set option %v to %v: %v", code, data, err)
		return
	}
	if _, ok := o.data[code]; ok {
		log.Printf("Not setting option %v more than once", code)
		return
	}
	o.order = append(o.order, code)
	o.data[code] = option
}

// Encode all options, including sentinel, to buf
func (o *Options) Encode(buf *bytes.Buffer) error {
	// Need the sentinel value at the end
	if len(o.order) > 0 && o.order[len(o.order)-1] != OPTION_SENTINEL {
		o.Set(OPTION_SENTINEL, nil)
	}

	for _, code := range o.order {
		option, ok := o.data[code]
		if !ok {
			log.Printf("Missing option %v ?", code)
			continue
		}

		// FIXME: why does the following fail to serialize?
		// binary.Write(buf, binary.LittleEndian, option)

		if err := buf.WriteByte(option.Header.Code); err != nil {
			return fmt.Errorf("Failed writing option code to buf: %v", err)
		}

		// If any of the following fail, we may generate badly corrupted data
		if err := buf.WriteByte(option.Header.Length); err != nil {
			return fmt.Errorf("Failed writing option length to buf: %v", err)
		}
		if len(option.Data) > 0 {
			if _, err := buf.Write(option.Data); err != nil {
				return fmt.Errorf("Failed writing option data to buf: %v", err)
			}
		}
	}
	return nil
}

// Parse options into a list
func ParseOptions(reader *bytes.Reader) *Options {
	options := NewOptions()
	for reader.Len() > 0 {
		option := &Option{}
		err := binary.Read(reader, binary.LittleEndian, &option.Header)
		if err != nil {
			log.Printf("Failed reading message option?")
			break
		}
		// Used for padding to word boundaries.
		if option.Header.Code == 0 {
			// !!!: Padding won't be followed by the length byte. Go back one.
			reader.Seek(-1, io.SeekCurrent)
			continue
		} else if option.Header.Code == OPTION_SENTINEL {
			// The end
			break
		}
		option.Data = make([]byte, option.Header.Length)
		count, err := reader.Read(option.Data)
		if err != nil {
			log.Printf("Failed reading: %v", err)
			break
		}
		if count != int(option.Header.Length) {
			log.Printf("Did not read as much as expected. %v != %v", count, option.Header.Length)
			break
		}
		options.Set(option.Header.Code, option.Data)
	}

	return options
}
