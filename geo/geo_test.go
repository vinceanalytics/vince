package geo

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	ip := net.ParseIP("81.2.69.142")
	got, err := Get(ip)
	require.NoError(t, err)
	require.Equal(t, Info{Country: "United Kingdom"}, got)
}
