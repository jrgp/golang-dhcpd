package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

type ConnectionHandler struct {
	remote         *net.UDPAddr
	request        *MessageHeader
	requestOptions *Options
	pool           *Pool
}

func NewConnectionHandler(message *DHCPMessage, remote *net.UDPAddr, pool *Pool) *ConnectionHandler {
	return &ConnectionHandler{
		remote:         remote,
		pool:           pool,
		request:        message.Header,
		requestOptions: message.Options,
	}
}

func (c *ConnectionHandler) Handle() {
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

	hostname := ""

	if option, ok := c.requestOptions.Get(OPTION_HOST_NAME); ok {
		hostname = string(option.Data)
	}

	mac := c.request.Mac
	log.Printf("DHCPDISCOVER from %v (%s)", mac.String(), hostname)
	if lease, ok := c.pool.GetLeaseByMac(mac); ok {
		log.Printf("Have old lease for %v: %v", mac.String(), lease.IP.String())
		c.SendLeaseInfo(lease, DHCPOFFER)
		return
	}

	lease, err := c.pool.GetNextLease(mac, hostname)
	if err != nil {
		log.Printf("Could not get a new lease for %v: %v", mac.String(), err)
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
	if lease, ok = c.pool.GetLeaseByMac(mac); !ok {
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
		ServerAddr: c.pool.MyIp,
		Mac:        c.request.Mac,
		Magic:      Magic,
	}

	log.Printf("Sending %s with %v to %v", opNames[op], lease.IP.String(), c.request.Mac.String())

	options := NewOptions()

	// Message type
	options.Set(OPTION_MESSAGE_TYPE, []byte{op})

	// Netmask option
	options.Set(OPTION_SUBNET, IpToFixedV4(c.pool.Netmask).Bytes())

	// Router (defgw)
	if len(c.pool.Router) > 0 {
		bytes := make([]byte, 0, 4*len(c.pool.Router))
		for _, ip := range c.pool.Router {
			bytes = append(bytes, ip...)
		}
		options.Set(OPTION_ROUTER, bytes)
	}

	// DNS servers
	if len(c.pool.Dns) > 0 {
		bytes := make([]byte, 0, 4*len(c.pool.Dns))
		for _, ip := range c.pool.Dns {
			bytes = append(bytes, ip...)
		}
		options.Set(OPTION_DNS_SERVER, bytes)
	}

	// Lease time
	options.Set(OPTION_LEASE_TIME, long2bytes(c.pool.LeaseTime))

	// DHCP server
	options.Set(OPTION_SERVER_ID, c.pool.MyIp.Bytes())

	buf := new(bytes.Buffer)

	err := header.Encode(buf)
	if err != nil {
		log.Printf("Writing dhcp header to our payload: %v", err)
		return
	}

	err = options.Encode(buf)
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
	dest := c.pool.Broadcast.String() + ":68"
	remote, err := net.ResolveUDPAddr("udp4", dest)
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
