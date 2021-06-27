package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

type ConnectionHandler struct {
	buf            []byte
	reader         *bytes.Reader
	remote         *net.UDPAddr
	request        *MessageHeader
	requestOptions []*MessageOption
	optionType     byte
	app            *App
}

func NewConnectionHandler(buf []byte, remote *net.UDPAddr, app *App) *ConnectionHandler {
	return &ConnectionHandler{
		buf:    buf,
		remote: remote,
		app:    app,
	}
}

func (c *ConnectionHandler) ParseRequest() error {
	if c.remote.Port != 68 {
		return fmt.Errorf("Source port is %d rather than 68", c.remote.Port)
	}
	c.request = &MessageHeader{}
	c.requestOptions = []*MessageOption{}
	c.reader = bytes.NewReader(c.buf)

	// Parse DHCP header
	err := binary.Read(c.reader, binary.LittleEndian, c.request)
	if err != nil {
		return fmt.Errorf("Failed unpacking into struct: %v", err)
	}

	// Verify sanity
	if c.request.HType != 1 {
		return fmt.Errorf("Only type 1 (ethernet) supported")
	}
	if c.request.HLen != 6 {
		return fmt.Errorf("Only 6 len mac addresses supported")
	}
	if c.request.Magic != Magic {
		return fmt.Errorf("Incorrect option magic")
	}

	// Parse arbitrary options
	for c.reader.Len() > 0 {
		option := &MessageOption{}
		err = binary.Read(c.reader, binary.LittleEndian, &option.Header)
		if err != nil {
			log.Printf("Failed reading message option?")
			break
		}
		// Used for padding to word boundaries. FIXME: padding won't be followed by length byte
		if option.Header.Code == 0 {
			continue
		} else if option.Header.Code == 255 {
			// The end
			break
		}
		option.Data = make([]byte, option.Header.Length)
		count, err := c.reader.Read(option.Data)
		if err != nil {
			log.Printf("Failed reading: %v", err)
			break
		}
		if count != int(option.Header.Length) {
			log.Printf("Did not read as much as expected.%v != %v", count, option.Header.Length)
			break
		}
		c.requestOptions = append(c.requestOptions, option)
		//log.Printf("Got option '%v': '%v' (%v)", option.Header.Code, option.Data, string(option.Data))

		// The op type can be specified as a dhcp option, and this should take
		// precedence
		if option.Header.Code == 53 && option.Header.Length == 1 {
			c.request.Op = option.Data[0]
		}

		// Similarly, ClientIP can be present here instead
		if option.Header.Code == 50 && option.Header.Length == 4 {
			c.request.ClientAddr = bytes2long(option.Data)
		}
	}

	return nil
}

func (c *ConnectionHandler) Handle() {
	if err := c.ParseRequest(); err != nil {
		log.Printf("Failed parsing request: %v", err)
		return
	}
	switch c.request.Op {
	case DHCPDISCOVER:
		c.HandleDiscover()
	case DHCPREQUEST:
		c.HandleRequest()
	default:
		log.Printf("Unimplemented op %v", c.request.Op)
	}
}

func (c *ConnectionHandler) HandleDiscover() {
	mac := c.request.Mac
	log.Printf("DHCPDISCOVER from %v", mac.String())
	if lease, ok := c.app.Pool.GetLeaseByMac(mac); ok {
		log.Printf("Have old lease for %v: %v", mac.String(), long2ip(lease.IP))
		c.SendOffer(lease)
		return
	}

	lease, err := c.app.Pool.GetNextLease(mac, "")
	if err != nil {
		log.Printf("Could not get a new lease for %v", mac.String())
		return
	}

	log.Printf("Got a new lease for %v: %v", mac.String(), long2ip(lease.IP))
	c.SendOffer(lease)
}

func (c *ConnectionHandler) HandleRequest() {
	mac := c.request.Mac
	log.Printf("DHCPREQUEST from %v", mac.String())
	var lease *Lease
	var ok bool
	if lease, ok = c.app.Pool.GetLeaseByMac(mac); !ok {
		log.Printf("Unrecognized lease for %v: %v", mac.String())
		return
	}

	// Verify IP matches what is in our lease
	if c.request.ClientAddr != lease.IP {
		log.Printf("Client IP does not match! %v != %v (expected)", c.request.ClientAddr, lease.IP)
		return
	}

	// Need to send

}

func (c *ConnectionHandler) SendOffer(lease *Lease) {
	header := &MessageHeader{
		Op:         DHCPOFFER,
		HType:      1,
		HLen:       6,
		Hops:       0,
		Identifier: c.request.Identifier,
		YourAddr:   lease.IP,
		ServerAddr: c.app.MyIp,
		Mac:        c.request.Mac,
		Magic:      Magic,
	}

	log.Printf("Sending DHCPOFFER with %v to %v", long2ip(lease.IP), c.request.Mac.String())

	options := []*MessageOption{}

	// Message type
	if option, err := NewMessageOption(53, []byte{DHCPOFFER}); err == nil {
		options = append(options, option)
	}

	// Netmask option
	if option, err := NewMessageOption(1, long2bytes(c.app.Pool.Mask)); err == nil {
		options = append(options, option)
	}

	// Router (defgw)
	if len(c.app.Pool.Router) > 0 {
		bytes := make([]byte, 0, 4*len(c.app.Pool.Router))
		for _, ip := range c.app.Pool.Router {
			bytes = append(bytes, long2bytes(ip)...)
		}
		if option, err := NewMessageOption(3, bytes); err == nil {
			options = append(options, option)
		}
	}

	// Lease time
	if option, err := NewMessageOption(51, long2bytes(c.app.Pool.LeaseTime)); err == nil {
		options = append(options, option)
	}

	// DHCP server
	if option, err := NewMessageOption(54, long2bytes(c.app.MyIp)); err == nil {
		options = append(options, option)
	}

	// DNS servers
	if len(c.app.Dns) > 0 {
		bytes := make([]byte, 0, 4*len(c.app.Dns))
		for _, ip := range c.app.Pool.Dns {
			bytes = append(bytes, long2bytes(ip)...)
		}
		if option, err := NewMessageOption(6, bytes); err == nil {
			options = append(options, option)
		}
	}

	// Sentinel
	if option, err := NewMessageOption(255, []byte{}); err == nil {
		options = append(options, option)
	}

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, header)
	if err != nil {
		log.Printf("Writing dhcp header to our payload: %v", err)
		return
	}

	for _, option := range options {
		// FIXME: why does the following fail to serialize?
		/*
			err = binary.Write(buf, binary.LittleEndian, option)
			if err != nil {
				log.Printf("Writing option %+v to our payload: %v", option, err)
				return
			}
		*/
		if err := buf.WriteByte(option.Header.Code); err != nil {
			log.Printf("Failed writing option code to buf: %v", err)
			return
		}
		if err := buf.WriteByte(option.Header.Length); err != nil {
			log.Printf("Failed writing option length to buf: %v", err)
			return
		}
		if len(option.Data) > 0 {
			if _, err := buf.Write(option.Data); err != nil {
				log.Printf("Failed writing option data to buf: %v", err)
				return
			}
		}
	}

	err = c.sendBroadcast(buf.Bytes())
	if err != nil {
		log.Printf("Failed sending DHCPOFFER payload: %v", err)
	}
}

func (c *ConnectionHandler) sendBroadcast(data []byte) error {
	// Quickly ripped from https://github.com/aler9/howto-udp-broadcast-golang
	local, err := net.ResolveUDPAddr("udp4", ":")
	if err != nil {
		return fmt.Errorf("Failed resolving local: %v", err)
	}
	remote, err := net.ResolveUDPAddr("udp4", "172.17.0.255:68") // FIXME: don't hardcode the address here
	if err != nil {
		return fmt.Errorf("Failed resolving remote: %v", err)
	}
	list, err := net.DialUDP("udp4", local, remote)
	if err != nil {
		return fmt.Errorf("Failed dialing: %v", err)
	}
	defer list.Close()
	_, err = list.Write(data)
	if err != nil {
		return fmt.Errorf("Failed writing: %v", err)
	}
	return nil
}
