package sets

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/roaring"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/mutex"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/tx"
)

func Rows(txn *tx.Tx, field string, start uint64, o *roaring64.Bitmap, filters ...roaring.BitmapFilter) error {
	return mutex.Rows(txn, field, start, o, filters...)
}
