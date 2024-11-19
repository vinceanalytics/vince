package ro2

import (
	"github.com/gernest/roaring"
)

type BitmapRowsUnion = roaring.BitmapRowsUnion

var NewBitmapRowsUnion = roaring.NewBitmapRowsUnion
var NewBitmapBitmapFilter = roaring.NewBitmapBitmapFilter

func Apply(ra *Bitmap, filter roaring.BitmapFilter) error {
	it, _ := ra.Containers.Iterator(0)
	return roaring.ApplyFilterToIterator(filter, it)
}
