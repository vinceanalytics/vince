package shards

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShardPath(t *testing.T) {
	got := shardPath("data", 1)
	require.Equal(t, filepath.Join("data", "000001"), got)
}
