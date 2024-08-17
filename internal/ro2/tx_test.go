package ro2

import (
	"slices"
	"testing"

	"github.com/dgraph-io/badger/v4"
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
		return tx.Add(0, 0, nil, nil, o)
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
		return tx.Add(0, 0, nil, nil, o)
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
		return tx.Add(0, 0, nil, nil, o)
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

func TestTranslations(t *testing.T) {
	db, err := newDB(t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	type T struct {
		shard, field uint64
		keys         []uint32
		values       []string
	}
	sample := []T{
		{0, 0, []uint32{1, 2, 3}, []string{"hello", "world", "jane"}},
		{0, 1, []uint32{1, 2, 3}, []string{"hello", "world", "jane"}}, // same field different shards
		{1, 0, []uint32{1, 2, 3}, []string{"hello", "world", "jane"}}, // different field same values
		{1, 1, []uint32{1, 2, 4}, []string{"hello", "world", "John"}}, // similar values
	}
	err = db.Update(func(tx *Tx) error {
		for _, s := range sample {
			err := tx.saveTranslations(s.shard, s.field, s.keys, s.values)
			if err != nil {
				return err
			}
		}
		return nil
	})
	require.NoError(t, err)

	t.Run("values are only stored once", func(t *testing.T) {
		var all []string
		db.db.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.IteratorOptions{
				Prefix: []byte{byte(TRANSLATE), 0},
			})
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				it.Item().Value(func(val []byte) error {
					all = append(all, string(val))
					return nil
				})
			}
			return nil
		})
		slices.Sort(all)
		require.Equal(t, []string{"John", "hello", "jane", "world"}, all)
	})

	t.Run("search", func(t *testing.T) {
		var got []T
		var want []T
		db.View(func(tx *Tx) error {
			for _, s := range sample {
				w := T{
					shard:  s.shard,
					field:  s.field,
					values: slices.Clone(s.values),
				}
				slices.Sort(w.values)
				g := T{
					shard: s.shard,
					field: s.field,
				}
				tx.searchTranslation(s.shard, s.field, func(val []byte) {
					g.values = append(g.values, string(val))
				})
				slices.Sort(g.values)
				want = append(want, w)
				got = append(got, g)
			}
			return nil
		})
		require.Equal(t, want, got)
	})

	t.Run("find", func(t *testing.T) {
		var got string
		db.View(func(tx *Tx) error {
			got = tx.Find(1)
			return nil
		})
		require.Equal(t, "hello", got)
	})
}
