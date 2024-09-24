package bsi

import (
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/tx"
	"github.com/gernest/roaring"
	"github.com/gernest/rows"
)

func SumCount(txn *tx.Tx, field string, o *roaring64.Bitmap, filters *rows.Row) (count int32, sum int64, err error) {
	err = txn.Cursor(field, func(c *rbf.Cursor, tx *tx.Tx) error {
		var filterData *roaring.Bitmap
		if filters != nil && len(filters.Segments) > 0 {
			filterData = filters.Segments[0].Data()
		}
		bsi := roaring.NewBitmapBSICountFilter(filterData)
		err := c.ApplyFilter(0, bsi)
		if err != nil {
			return err
		}
		count, sum = bsi.Total()
		return nil
	})
	return
}
