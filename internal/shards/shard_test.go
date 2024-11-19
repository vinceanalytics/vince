package shards

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	one := Format(1)
	require.Equal(t, "000001", one)
	require.Equal(t, uint64(1), Parse(one))
}
