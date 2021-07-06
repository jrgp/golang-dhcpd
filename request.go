package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

type RequestHandler struct {
	header  *MessageHeader
	options *Options
	pool    *Pool
}

func NewRequestHandler(message *DHCPMessage, pool *Pool) *RequestHandler {
	return &RequestHandler{
		pool:    pool,
		header:  message.Header,
		options: message.Options,
	}
}

func (r *RequestHandler) Handle() *DHCPMessage {
	switch r.header.Op {
	case DHCPDISCOVER:
		return r.HandleDiscover()
	case DHCPREQUEST:
		return r.HandleRequest()
	case DHCPRELEASE:
		return r.HandleRelease()
	default:
		log.Printf("Unimplemented op %v", r.header.Op)
		return nil
	}
}

func (r *RequestHandler) HandleDiscover() *DHCPMessage {
	hostname := ""

	if option, ok := r.options.Get(OPTION_HOST_NAME); ok {
		hostname = string(option.Data)
	}

	mac := r.header.Mac
	log.Printf("DHCPDISCOVER from %v (%s)", mac.String(), hostname)
	if lease, ok := r.pool.TouchLeaseByMac(mac); ok {
		log.Printf("Have old lease for %v: %v", mac.String(), lease.IP.String())
		return r.SendLeaseInfo(lease, DHCPOFFER)
	}

	lease, err := r.pool.GetNextLease(mac, hostname)
	if err != nil {
		log.Printf("Could not get a new lease for %v: %v", mac.String(), err)
		return nil
	}

	return r.SendLeaseInfo(lease, DHCPOFFER)
}

func (r *RequestHandler) HandleRequest() *DHCPMessage {
	mac := r.header.Mac
	log.Printf("DHCPREQUEST from %v for %v", mac.String(), r.header.ClientAddr.String())
	var lease *Lease
	var ok bool
	if lease, ok = r.pool.TouchLeaseByMac(mac); !ok {
		log.Printf("Unrecognized lease for %v", mac.String())
		return r.SendNAK()

	}

	// Verify IP matches what is in our lease
	if r.header.ClientAddr != lease.IP {
		log.Printf("Client IP does not match! %v != %v (expected)", r.header.ClientAddr, lease.IP)
		return r.SendNAK()
	}

	// Need to send DHCPACK
	return r.SendLeaseInfo(lease, DHCPACK)
}

func (r *RequestHandler) HandleRelease() *DHCPMessage {
	mac := r.header.Mac

	log.Printf("DHCPRELEASE from %v for %v", mac.String(), r.header.ClientAddr.String())
	var lease *Lease
	var ok bool

	if lease, ok = r.pool.ReleaseLeaseByMac(mac); !ok {
		log.Printf("Unrecognized lease for %v to release", mac.String())
		return nil
	}

	// Verify IP matches what is in our lease
	if r.header.ClientAddr != lease.IP {
		log.Printf("Client IP does not match! %v != %v (expected)", r.header.ClientAddr, lease.IP)
	}

	// No response to a DHCPRELEASE
	return nil
}

// Share code for DHCPOFFER and DHCPACK
func (r *RequestHandler) SendLeaseInfo(lease *Lease, op byte) *DHCPMessage {
	header := &MessageHeader{
		Op:         op,
		Hops:       0,
		Identifier: r.header.Identifier,
		YourAddr:   lease.IP,
		ServerAddr: r.pool.MyIp,
		Mac:        r.header.Mac,
	}

	log.Printf("Sending %s with %v to %v", opNames[op], lease.IP.String(), r.header.Mac.String())

	options := NewOptions()

	// Message type
	options.Set(OPTION_MESSAGE_TYPE, []byte{op})

	// Netmask option
	options.Set(OPTION_SUBNET, IpToFixedV4(r.pool.Netmask).Bytes())

	// Router (defgw)
	if len(r.pool.Router) > 0 {
		bytes := make([]byte, 0, 4*len(r.pool.Router))
		for _, ip := range r.pool.Router {
			bytes = append(bytes, ip...)
		}
		options.Set(OPTION_ROUTER, bytes)
	}

	// DNS servers
	if len(r.pool.Dns) > 0 {
		bytes := make([]byte, 0, 4*len(r.pool.Dns))
		for _, ip := range r.pool.Dns {
			bytes = append(bytes, ip...)
		}
		options.Set(OPTION_DNS_SERVER, bytes)
	}

	// Lease time
	options.Set(OPTION_LEASE_TIME, long2bytes(uint32(r.pool.LeaseTime.Seconds())))

	// DHCP server
	options.Set(OPTION_SERVER_ID, r.pool.MyIp.Bytes())

	return &DHCPMessage{header, options}
}

func (r *RequestHandler) SendNAK() *DHCPMessage {
	header := &MessageHeader{
		Op:         DHCPNAK,
		Hops:       0,
		Identifier: r.header.Identifier,
		ServerAddr: r.pool.MyIp,
		Mac:        r.header.Mac,
	}

	log.Printf("Sending %s to %v", opNames[header.Op], r.header.Mac.String())

	options := NewOptions()
	options.Set(OPTION_MESSAGE_TYPE, []byte{header.Op})

	// FIXME: we likely need more options

	return &DHCPMessage{header, options}
}

func (r *RequestHandler) sendMessageBroadcast(message *DHCPMessage) {
	buf := new(bytes.Buffer)

	err := message.Encode(buf)
	if err != nil {
		log.Printf("Failed encoding payload: %v", err)
		return
	}

	err = r.sendBroadcast(buf.Bytes())
	if err != nil {
		log.Printf("Failed sending %s payload: %v", opNames[message.Header.Op], err)
	}
}

func (r *RequestHandler) sendBroadcast(data []byte) error {
	// Quickly ripped from https://github.com/aler9/howto-udp-broadcast-golang
	local, err := net.ResolveUDPAddr("udp4", ":")
	if err != nil {
		return fmt.Errorf("Failed resolving local: %v", err)
	}
	dest := r.pool.Broadcast.String() + ":68"
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
