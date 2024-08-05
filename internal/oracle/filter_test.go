package oracle

import (
	"testing"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/cursor"
)

func TestFilter(t *testing.T) {
	db := rbf.NewDB(t.TempDir(), nil)
	db.Open()
	defer db.Close()

	tx, _ := db.Begin(true)
	defer tx.Rollback()

	b := roaring.NewBitmap()
	for id := range uint64(10) {
		b.Add((id+20)*shardwidth.ShardWidth + (id % shardwidth.ShardWidth))
	}
	tx.AddRoaring("test", b)

	fs := roaring.NewBitmapRowsUnion([]uint64{3 + 20})
	tx.ApplyFilter("test", 0, fs)

	r := rows.NewRowFromBitmap(fs.Results(0))

	require.Equal(t, []uint64{3}, r.Columns())

	var min, max uint64
	cursor.Tx(tx, "test", func(c *rbf.Cursor) error {
		min, _, _ = cursor.MinRowID(c)
		max, _ = cursor.MaxRowID(c)
		return nil
	})
	require.Equal(t, []uint64{20, 29}, []uint64{min, max})
}

func TestMinMax_ts(t *testing.T) {
	db := rbf.NewDB(t.TempDir(), nil)
	db.Open()
	defer db.Close()

	tx, _ := db.Begin(true)
	defer tx.Rollback()

	b := roaring.NewBitmap()
	for id := range uint64(10) {
		setValue(b, id, int64(id)*10)
	}
	tx.AddRoaring("test", b)

	var min, max int64
	cursor.Tx(tx, "test", func(c *rbf.Cursor) error {
		var err error
		min, max, err = MinMax(c, 0)
		return err
	})
	require.Equal(t, []int64{0, 90}, []int64{min, max})
}
