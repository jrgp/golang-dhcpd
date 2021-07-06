package main

import (
	"github.com/stretchr/testify/require"

	"net"
	"testing"
)

func TestIpIntConversions(t *testing.T) {
	require.Equal(t, uint32(167772161), ip2long(net.ParseIP("10.0.0.1")))
	require.Equal(t, uint32(167772162), ip2long(net.ParseIP("10.0.0.2")))

	require.Equal(t, net.ParseIP("10.0.0.1").To4(), long2ip(167772161))
	require.Equal(t, net.ParseIP("10.0.0.2").To4(), long2ip(167772162))

	require.Equal(t, net.ParseIP("10.0.0.2").To4(), long2ip(ip2long(net.ParseIP("10.0.0.2").To4())))
}

func TestCalcBroadcast(t *testing.T) {
	require.Equal(t, net.ParseIP("10.0.0.255").To4(), calcBroadcast(net.ParseIP("10.0.0.0"), net.ParseIP("255.255.255.0")).To4())
	require.Equal(t, net.ParseIP("172.17.0.255").To4(), calcBroadcast(net.ParseIP("172.17.0.0"), net.ParseIP("255.255.255.0")).To4())
}
