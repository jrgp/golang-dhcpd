package main

import (
	"github.com/stretchr/testify/require"

	"net"
	"testing"
)

func TestParseDhcpMessage(t *testing.T) {
	b := []byte{
		1, 1, 6, 0, 110, 255, 201, 48, 0, 3, 0, 0, 172, 17, 0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 28, 66, 180, 110, 29, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 99, 130, 83, 99, 53, 1, 3, 12, 7, 117, 98, 117, 110, 116, 117, 50, 55, 13, 1, 28, 2, 3, 15, 6, 119, 12, 44, 47, 26, 121, 42, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	message, err := ParseDhcpMessage(b)
	require.Nil(t, err)

	require.Equal(t, "0:1c:42:b4:6e:1d", message.Header.Mac.String())
	require.Equal(t, byte(DHCPREQUEST), message.Header.Op)
	require.Equal(t, uint32(818544494), message.Header.Identifier)
	require.Equal(t, IpToFixedV4(net.ParseIP("172.17.0.100")), message.Header.ClientAddr)

	opt, ok := message.Options.Get(OPTION_MESSAGE_TYPE)
	require.True(t, ok)
	require.Equal(t, []byte{3}, opt.Data)

	opt, ok = message.Options.Get(OPTION_HOST_NAME)
	require.True(t, ok)
	require.Equal(t, []byte("ubuntu2"), opt.Data)

	// Verify added padding around options doesn't break
	b = []byte{
		1, 1, 6, 0, 110, 255, 201, 48, 0, 3, 0, 0, 172, 17, 0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 28, 66, 180, 110, 29, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 99, 130, 83, 99, 53, 1, 3, 0, 0, 0, 0, 0, 12, 7, 117, 98, 117, 110, 116, 117, 50, 55, 13, 1, 28, 2, 3, 15, 6, 119, 12, 44, 47, 26, 121, 42, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	message, err = ParseDhcpMessage(b)
	require.Nil(t, err)

	opt, ok = message.Options.Get(OPTION_HOST_NAME)
	require.True(t, ok)
	require.Equal(t, []byte("ubuntu2"), opt.Data)

	// Trying to decode something broken fails
	b = []byte{}
	_, err = ParseDhcpMessage(b)
	require.NotNil(t, err)
}
