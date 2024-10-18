package roaring

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOffsetRange(t *testing.T) {
	ra := NewBitmap()
	id := shardWidth + 4
	ra.Mutex(uint64(id), 5)
	o := ra.Row(1, 5)
	require.Equal(t, []uint64{shardWidth + 4}, o.ToArray())
}
