package main

import (
	"github.com/stretchr/testify/require"

	"net"
	"testing"
)

func TestCalcBroadcast(t *testing.T) {
	require.Equal(t, net.ParseIP("10.0.0.255").To4(), calcBroadcast(net.ParseIP("10.0.0.0"), net.ParseIP("255.255.255.0")).To4())
	require.Equal(t, net.ParseIP("172.17.0.255").To4(), calcBroadcast(net.ParseIP("172.17.0.0"), net.ParseIP("255.255.255.0")).To4())
}
