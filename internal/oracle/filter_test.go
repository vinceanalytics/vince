package oracle

import (
	"testing"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/rbf"
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
}
