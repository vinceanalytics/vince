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

func TestExtractMutex(t *testing.T) {
	b := NewBitmap()
	b.Mutex(1, 4)
	b.Mutex(2, 6)
	b.Mutex(4, 4)

	m := NewBitmap()
	m.Set(1)
	m.Set(4)
	want := map[uint64][]uint64{
		4: {1, 4},
	}
	got := map[uint64][]uint64{}
	b.ExtractMutex(m, func(row uint64, columns *Bitmap) error {
		got[row] = columns.ToArray()
		return nil
	})
	require.Equal(t, want, got)
}

func TestSumBSI(t *testing.T) {

	sample := []struct {
		ids    []uint64
		values []int64
		total  int64
	}{
		{
			[]uint64{1, 2, 3},
			[]int64{1, 2, 3},
			6,
		},
		{
			[]uint64{1, 2, 3},
			[]int64{1, -1, -1},
			-1,
		},
	}
	for _, s := range sample {
		b := NewBitmap()
		m := NewBitmap()
		for i := range s.ids {
			b.BSI(s.ids[i], s.values[i])
			m.Set(s.ids[i])
		}
		require.Equal(t, s.total, b.BSISum(0, m))
	}
}
