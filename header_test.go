package main

import (
	"github.com/stretchr/testify/require"

	"bytes"
	"testing"
)

func TestMessageParseEncode(t *testing.T) {
	b := []byte{
		1, 1, 6, 0, 110, 255, 201, 48, 0, 3, 0, 0, 172, 17, 0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 28, 66, 180, 110, 29, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 99, 130, 83, 99, 53, 1, 3, 12, 7, 117, 98, 117, 110, 116, 117, 50, 55, 13, 1, 28, 2, 3, 15, 6, 119, 12, 44, 47, 26, 121, 42, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}

	reader := bytes.NewReader(b)

	header, err := ParseMessageHeader(reader)
	require.Nil(t, err)

	// Break things slightly to ensure .Encode fixes them, and that we don't
	// need to set them manually when creating messages
	header.Magic = 0
	header.HLen = 0
	header.HType = 0

	buf := new(bytes.Buffer)
	err = header.Encode(buf)
	require.Nil(t, err)
	encoded := buf.Bytes()

	require.Equal(t, b[:240], encoded)
}

func TestMacEncoding(t *testing.T) {
	encoded := StrToMac("0:1c:42:b4:6e:1d")
	require.Equal(t, "0:1c:42:b4:6e:1d", encoded.String())
	require.Equal(t, MacAddress{0, 0x1c, 0x42, 0xb4, 0x6e, 0x1d}, encoded)
}
