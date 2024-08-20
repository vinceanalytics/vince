package ro2

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func TestTxAdd(t *testing.T) {
	db, err := newDB(t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	s := []struct {
		id, value uint64
	}{
		{0, 12},
		{1, 18},
		{2, 12},
		{20, 18},
	}
	o := roaring64.New()
	for i := range s {
		o.Add(ro.MutexPosition(s[i].id, s[i].value))
	}
	err = db.Update(func(tx *Tx) error {
		return tx.Add(0, 0, o)
	})
	require.NoError(t, err)
	var result *roaring64.Bitmap
	match := map[uint64][]uint16{}
	db.View(func(tx *Tx) error {
		result = tx.Row(0, 0, 12)
		tx.ExtractMutex(0, 0, result, func(row uint64, c *roaring.Container) {
			c.Each(func(u uint16) bool {
				match[row] = append(match[row], u)
				return true
			})
		})
		return nil
	})
	require.Equal(t, []uint64{0, 2}, result.ToArray())
	require.Equal(t, map[uint64][]uint16{
		12: {0, 2},
	}, match)
}

func TestTxAdd_bsi(t *testing.T) {
	db, err := newDB(t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	s := []struct {
		id    uint64
		value int64
	}{
		{0, 12},
		{1, 18},
		{2, 12},
		{20, -18},
	}
	o := roaring64.New()
	for i := range s {
		ro.BSI(o, s[i].id, s[i].value)
	}
	err = db.Update(func(tx *Tx) error {
		return tx.Add(0, 0, o)
	})
	require.NoError(t, err)

	match := map[uint64]int64{}
	var maxKey uint64
	db.View(func(tx *Tx) error {
		result := roaring64.New()
		result.AddMany([]uint64{1, 20})
		tx.ExtractBSI(0, 0, result, func(row uint64, v int64) {
			match[row] = v
		})
		maxKey, _ = tx.max(0, 0)
		return nil
	})
	require.Equal(t, map[uint64]int64{
		1:  18,
		20: -18,
	}, match)
	require.Equal(t, o.Maximum(), maxKey)
}

func TestTxCmp_range(t *testing.T) {
	db, err := newDB(t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	s := []struct {
		id    uint64
		value int64
	}{
		{0, 12},
		{1, 13},
		{2, 14},
		{20, 15},
		{22, 16},
		{23, 17},
	}
	o := roaring64.New()
	for i := range s {
		ro.BSI(o, s[i].id, s[i].value)
	}
	err = db.Update(func(tx *Tx) error {
		return tx.Add(0, 0, o)
	})
	require.NoError(t, err)

	var match []uint64
	db.View(func(tx *Tx) error {
		b := tx.Cmp(0, 0, roaring64.RANGE, 14, 16)
		match = b.ToArray()
		return nil
	})
	require.Equal(t, []uint64{2, 20}, match)
}
