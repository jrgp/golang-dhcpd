package main

import (
	"github.com/stretchr/testify/require"

	"net"
	"testing"
)

func TestDhcpDiscover(t *testing.T) {
	pool := NewPool()
	pool.Start = net.ParseIP("10.0.0.10")
	pool.End = net.ParseIP("10.0.0.20")
	pool.Netmask = net.ParseIP("255.255.255.0")
	pool.Router = []net.IP{net.ParseIP("10.0.0.1")}
	pool.MyIp = IpToFixedV4(net.ParseIP("10.0.0.254"))
	pool.Dns = []net.IP{net.ParseIP("1.1.1.1"), net.ParseIP("1.0.0.1")}

	//
	// DHCPREQUEST targeting an unknown lease. Should get a NAK back
	//

	// FIXME: consider creating DhcpMessage{} directly instead of these giant blobs
	b := []byte{
		1, 1, 6, 0, 110, 255, 201, 48, 0, 3, 0, 0, 172, 17, 0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 28, 66, 180, 110, 29, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 99, 130, 83, 99, 53, 1, 3, 12, 7, 117, 98, 117, 110, 116, 117, 50, 55, 13, 1, 28, 2, 3, 15, 6, 119, 12, 44, 47, 26, 121, 42, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	message, err := ParseDhcpMessage(b)
	require.Nil(t, err)

	handler := NewRequestHandler(message, pool)
	response := handler.Handle()

	require.Equal(t, BOOT_REPLY, response.Header.Op)
	require.Equal(t, DHCPNAK, response.Options.GetByte(OPTION_MESSAGE_TYPE))

	//
	// DISCOVER. Should get back a lease.
	//
	b = []byte{
		1, 1, 6, 0, 237, 92, 70, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 28, 66, 180, 110, 29, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 99, 130, 83, 99, 53, 1, 1, 12, 7, 117, 98, 117, 110, 116, 117, 50, 55, 13, 1, 28, 2, 3, 15, 6, 119, 12, 44, 47, 26, 121, 42, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	message, err = ParseDhcpMessage(b)
	require.Nil(t, err)
	require.Equal(t, DHCPDISCOVER, message.Options.GetByte(OPTION_MESSAGE_TYPE))

	handler = NewRequestHandler(message, pool)
	response = handler.Handle()

	require.Equal(t, BOOT_REPLY, response.Header.Op)
	require.Equal(t, DHCPOFFER, response.Options.GetByte(OPTION_MESSAGE_TYPE))
	require.Equal(t, IpToFixedV4(net.ParseIP("10.0.0.10")), response.Header.YourAddr)

	require.Equal(t, []FixedV4{IpToFixedV4(net.ParseIP("255.255.255.0"))}, response.Options.GetFixedV4s(OPTION_SUBNET))
	require.Equal(t, []FixedV4{IpToFixedV4(net.ParseIP("10.0.0.1"))}, response.Options.GetFixedV4s(OPTION_ROUTER))
	require.Equal(t, []FixedV4{IpToFixedV4(net.ParseIP("1.1.1.1")), IpToFixedV4(net.ParseIP("1.0.0.1"))}, response.Options.GetFixedV4s(OPTION_DNS_SERVER))
	require.Equal(t, []FixedV4{IpToFixedV4(net.ParseIP("10.0.0.254"))}, response.Options.GetFixedV4s(OPTION_SERVER_ID))

	// Pool should have a lease for this mac
	lease, ok := pool.TouchLeaseByMac(message.Header.Mac)
	require.True(t, ok)
	require.Equal(t, IpToFixedV4(net.ParseIP("10.0.0.10")), lease.IP)

	//
	// Request targeting the wrong IP should get back a NAK
	//
	b = []byte{
		1, 1, 6, 0, 110, 255, 201, 48, 0, 3, 0, 0, 10, 0, 0, 11, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 28, 66, 180, 110, 29, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 99, 130, 83, 99, 53, 1, 3, 12, 7, 117, 98, 117, 110, 116, 117, 50, 55, 13, 1, 28, 2, 3, 15, 6, 119, 12, 44, 47, 26, 121, 42, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	message, err = ParseDhcpMessage(b)
	require.Nil(t, err)
	require.Equal(t, BOOT_REPLY, response.Header.Op)
	require.Equal(t, DHCPREQUEST, message.Options.GetByte(OPTION_MESSAGE_TYPE))

	handler = NewRequestHandler(message, pool)
	response = handler.Handle()

	require.Equal(t, DHCPNAK, response.Options.GetByte(OPTION_MESSAGE_TYPE))

	//
	// A request should now get back an ACK
	//
	b = []byte{
		1, 1, 6, 0, 110, 255, 201, 48, 0, 3, 0, 0, 10, 0, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 28, 66, 180, 110, 29, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 99, 130, 83, 99, 53, 1, 3, 12, 7, 117, 98, 117, 110, 116, 117, 50, 55, 13, 1, 28, 2, 3, 15, 6, 119, 12, 44, 47, 26, 121, 42, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	message, err = ParseDhcpMessage(b)
	require.Nil(t, err)
	require.Equal(t, DHCPREQUEST, message.Options.GetByte(OPTION_MESSAGE_TYPE))

	handler = NewRequestHandler(message, pool)
	response = handler.Handle()

	require.Equal(t, BOOT_REPLY, response.Header.Op)
	require.Equal(t, DHCPACK, response.Options.GetByte(OPTION_MESSAGE_TYPE))

	//
	// Do a DHCPRELEASE
	//
	b = []byte{
		1, 1, 6, 0, 245, 234, 140, 40, 0, 0, 0, 0, 10, 0, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 28, 66, 180, 110, 29, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 99, 130, 83, 99, 53, 1, 7, 54, 4, 172, 17, 0, 1, 12, 7, 117, 98, 117, 110, 116, 117, 50, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	message, err = ParseDhcpMessage(b)
	require.Nil(t, err)
	require.Equal(t, DHCPRELEASE, message.Options.GetByte(OPTION_MESSAGE_TYPE))

	handler = NewRequestHandler(message, pool)
	response = handler.Handle()

	require.Nil(t, response)

	// Pool should no longer have a lease for this mac
	lease, ok = pool.TouchLeaseByMac(message.Header.Mac)
	require.False(t, ok)
	require.Nil(t, lease)
}
