package main

import (
	"github.com/stretchr/testify/require"

	"net"
	"testing"
	"time"
)

func TestIpAllocation(t *testing.T) {
	// Pool with deliberately only 2 available IPs
	pool := NewPool()
	pool.Start = net.ParseIP("172.0.0.10")
	pool.End = net.ParseIP("172.0.0.11")
	pool.Netmask = net.ParseIP("255.255.255.0")
	pool.LeaseTime = time.Duration(1) * time.Hour

	mac1 := MacAddress{0, 0, 0, 0, 0, 1}
	mac2 := MacAddress{0, 0, 0, 0, 0, 2}
	mac3 := MacAddress{0, 0, 0, 0, 0, 3}

	// Verify initial IP lease acquisition works
	lease1, err := pool.GetNextLease(mac1, "host1")
	require.Nil(t, err)
	require.Equal(t, IpToFixedV4(net.ParseIP("172.0.0.10")), lease1.IP)
	require.Equal(t, mac1, lease1.Mac)
	require.Equal(t, "host1", lease1.Hostname)
	require.False(t, lease1.Expired())
	orig_time := lease1.Expiration

	// And that when we bump it, its expiration gets bumped accordingly
	lease1Fetched, ok := pool.TouchLeaseByMac(mac1)
	require.True(t, ok)
	require.True(t, lease1Fetched.Expiration.After(orig_time))

	// And that another host is able to get the next free IP
	lease2, err := pool.GetNextLease(mac2, "host2")
	require.Nil(t, err)
	require.Equal(t, mac2, lease2.Mac)
	require.Equal(t, "host2", lease2.Hostname)
	require.Equal(t, IpToFixedV4(net.ParseIP("172.0.0.11")), lease2.IP)
	require.False(t, lease2.Expired())

	// No free Ips for lease3 so it will fail
	lease3, err := pool.GetNextLease(mac3, "host3")
	require.Equal(t, ErrNoIps, err)
	require.Nil(t, lease3)

	// However, if we expire lease1, host3 will get its IP
	lease1.Expiration = time.Now().Add(time.Duration(-1) * time.Hour)
	require.True(t, lease1.Expired())

	lease3, err = pool.GetNextLease(mac3, "host3")
	require.Nil(t, err)
	require.Equal(t, mac3, lease3.Mac)
	require.Equal(t, "host3", lease3.Hostname)
	require.Equal(t, IpToFixedV4(net.ParseIP("172.0.0.10")), lease3.IP)
}
