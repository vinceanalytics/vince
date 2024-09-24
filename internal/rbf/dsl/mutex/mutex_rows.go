package mutex

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/roaring"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/cursor"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/tx"
)

func Rows(txn *tx.Tx, field string, start uint64, o *roaring64.Bitmap, filters ...roaring.BitmapFilter) error {
	return txn.Cursor(field, func(c *rbf.Cursor, tx *tx.Tx) error {
		return cursor.Rows(c, start, func(row uint64) error {
			o.Add(row)
			return nil
		}, filters...)
	})
}
