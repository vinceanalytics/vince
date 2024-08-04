package oracle

import (
	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/cursor"
	"go.etcd.io/bbolt"
)

func (o *Oracle) CurrentVisitors(start, end int64, domain string, filter Filter) (uint64, error) {

	// Visitors are counted as unique uid. But we chunk data across shard, there is
	// a possibility a uid is observed in multiple shards.
	//
	// Here a map[uint64]struct{} should suffice , but since we already rely on
	// roaring bitmaps we can just use one here to track all unique uid across shards.
	bsi := roaring64.NewDefaultBSI()

	err := o.db.Select(start, end, domain, Noop(), func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, match *rows.Row) error {
		return cursor.Tx(rTx, "uid", func(c *rbf.Cursor) error {
			return extractBSI(c, shard, match, func(column uint64, value int64) error {
				bsi.SetValue(column, value)
				return nil
			})
		})
	})
	if err != nil {
		return 0, err
	}
	return bsi.Transpose().GetCardinality(), nil
}
