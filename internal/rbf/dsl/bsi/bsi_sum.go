package bsi

import (
	"github.com/gernest/roaring"
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
)

func Sum(cu *rbf.Cursor, filters *rows.Row, f func(count int32, sum int64) error) error {
	var filterData *roaring.Bitmap
	if filters != nil && len(filters.Segments) > 0 {
		filterData = filters.Segments[0].Data()
	}
	bsi := roaring.NewBitmapBSICountFilter(filterData)
	err := cu.ApplyFilter(0, bsi)
	if err != nil {
		return err
	}
	return f(bsi.Total())
}
