package mutex

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/cursor"
	"github.com/gernest/rbf/dsl/tx"
	"github.com/gernest/roaring"
)

func Rows(txn *tx.Tx, field string, start uint64, o *roaring64.Bitmap, filters ...roaring.BitmapFilter) error {
	return txn.Cursor(field, func(c *rbf.Cursor, tx *tx.Tx) error {
		return cursor.Rows(c, start, func(row uint64) error {
			o.Add(row)
			return nil
		}, filters...)
	})
}
