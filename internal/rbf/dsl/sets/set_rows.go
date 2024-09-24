package sets

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf/dsl/mutex"
	"github.com/gernest/rbf/dsl/tx"
	"github.com/gernest/roaring"
)

func Rows(txn *tx.Tx, field string, start uint64, o *roaring64.Bitmap, filters ...roaring.BitmapFilter) error {
	return mutex.Rows(txn, field, start, o, filters...)
}
