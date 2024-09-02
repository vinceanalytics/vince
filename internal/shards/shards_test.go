package shards

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShards(t *testing.T) {
	sx := New()
	sx.Set(
		[]uint32{1, 2, 3, 4, 5, 6},
		[]int64{0, 2, 4, 6, 8, 10},
	)
	t.Run("exact ts same range", func(t *testing.T) {
		require.Equal(t, []uint64{1}, sx.Select(0, 0))
		require.Equal(t, []uint64{6}, sx.Select(10, 10))
	})
	t.Run("not exact ts same range", func(t *testing.T) {
		require.Equal(t, []uint64{1}, sx.Select(1, 1))
		require.Equal(t, []uint64{6}, sx.Select(11, 11))
	})
	t.Run("small range", func(t *testing.T) {
		require.Equal(t, []uint64{1}, sx.Select(1, 1))
		require.Equal(t, []uint64{1}, sx.Select(1, 2))
		require.Equal(t, []uint64{6}, sx.Select(10, 11))
	})
	t.Run("big range", func(t *testing.T) {
		require.Equal(t, []uint64{2, 3, 4, 5}, sx.Select(2, 10))
	})
}
