package main

import (
	"golang.org/x/net/ipv4"

	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

type ConnectionHandler struct {
	buf, oob       []byte
	reader         *bytes.Reader
	remote         *net.UDPAddr
	request        *MessageHeader
	requestOptions *Options
	optionType     byte
	app            *App
}

func NewConnectionHandler(buf, oob []byte, remote *net.UDPAddr, app *App) *ConnectionHandler {
	return &ConnectionHandler{
		buf:    buf,
		oob:    oob,
		remote: remote,
		app:    app,
	}
}

func (c *ConnectionHandler) ParseRequest() error {
	if c.remote.Port != 68 {
		return fmt.Errorf("Source port is %d rather than 68", c.remote.Port)
	}
	c.request = &MessageHeader{}
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
	c.requestOptions = ParseOptions(c.reader)

	// Confusingly, the Op type can be overridden using an option
	if option, ok := c.requestOptions.Get(OPTION_MESSAGE_TYPE); ok {
		if option.Header.Length == 1 {
			c.request.Op = option.Data[0]
		}
	}

	// Similarly, so can the ClientAddr
	if option, ok := c.requestOptions.Get(OPTION_REQUESTED_IP); ok {
		if option.Header.Length == 4 {
			ip, err := BytesToFixedV4(option.Data)
			if err == nil {
				c.request.ClientAddr = ip
			} else {
				log.Printf("Failed converting byte stream to fixed v4")
			}
		}
	}

	return nil
}

func (c *ConnectionHandler) Handle() {
	cm := &ipv4.ControlMessage{}
	if err := cm.Parse(c.oob); err == nil {
		if cm.IfIndex != c.app.Pool.Nic.Index {
			log.Printf("Ignoring dhcp traffic on other interface")
			return
		}
	} else {
		log.Printf("failed parsing oob: %v", err)
		return
	}

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
		log.Printf("Have old lease for %v: %v", mac.String(), lease.IP.String())
		c.SendLeaseInfo(lease, DHCPOFFER)
		return
	}

	lease, err := c.app.Pool.GetNextLease(mac, "")
	if err != nil {
		log.Printf("Could not get a new lease for %v", mac.String())
		return
	}

	log.Printf("Got a new lease for %v: %v", mac.String(), lease.IP.String())
	c.SendLeaseInfo(lease, DHCPOFFER)
}

func (c *ConnectionHandler) HandleRequest() {
	mac := c.request.Mac
	log.Printf("DHCPREQUEST from %v", mac.String())
	var lease *Lease
	var ok bool
	if lease, ok = c.app.Pool.GetLeaseByMac(mac); !ok {
		// FIXME: handle this gracefully
		log.Printf("Unrecognized lease for %v. Rebranding as discover.", mac.String())
		c.HandleDiscover()
		return
	}

	// Verify IP matches what is in our lease
	if c.request.ClientAddr != lease.IP {
		// FIXME: handle this gracefully
		log.Printf("Client IP does not match! %v != %v (expected)", c.request.ClientAddr, lease.IP)
		return
	}

	// Need to send DHCPACK
	c.SendLeaseInfo(lease, DHCPACK)
}

// Share code for DHCPOFFER and DHCPACK
func (c *ConnectionHandler) SendLeaseInfo(lease *Lease, op byte) {
	header := &MessageHeader{
		Op:         op,
		HType:      1,
		HLen:       6,
		Hops:       0,
		Identifier: c.request.Identifier,
		YourAddr:   lease.IP,
		ServerAddr: c.app.MyIp,
		Mac:        c.request.Mac,
		Magic:      Magic,
	}

	log.Printf("Sending %s with %v to %v", opNames[op], lease.IP.String(), c.request.Mac.String())

	options := NewOptions()

	// Message type
	options.Set(OPTION_MESSAGE_TYPE, []byte{op})

	// Netmask option
	options.Set(OPTION_SUBNET, IpToFixedV4(c.app.Pool.Mask).Bytes())

	// Router (defgw)
	if len(c.app.Pool.Router) > 0 {
		bytes := make([]byte, 0, 4*len(c.app.Pool.Router))
		for _, ip := range c.app.Pool.Router {
			bytes = append(bytes, ip...)
		}
		options.Set(OPTION_ROUTER, bytes)
	}

	// DNS servers
	if len(c.app.Dns) > 0 {
		bytes := make([]byte, 0, 4*len(c.app.Pool.Dns))
		for _, ip := range c.app.Pool.Dns {
			bytes = append(bytes, ip...)
		}
		options.Set(OPTION_DNS_SERVER, bytes)
	}

	// Lease time
	options.Set(OPTION_LEASE_TIME, long2bytes(c.app.Pool.LeaseTime))

	// DHCP server
	options.Set(OPTION_SERVER_ID, c.app.MyIp.Bytes())

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, header)
	if err != nil {
		log.Printf("Writing dhcp header to our payload: %v", err)
		return
	}

	_, err = buf.Write(options.Encode())
	if err != nil {
		log.Printf("Writing dhcp options to our payload: %v", err)
		return
	}

	err = c.sendBroadcast(buf.Bytes())
	if err != nil {
		log.Printf("Failed sending %s payload: %v", opNames[op], err)
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
