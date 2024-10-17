package bsi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKV_BitCount(t *testing.T) {
	require.Equal(t, 0, KV{}.BitCount())
	require.Equal(t, 0, make(KV, 1).BitCount())
	require.Equal(t, 1, make(KV, 2).BitCount())
	require.Equal(t, 2, make(KV, 3).BitCount())
}
