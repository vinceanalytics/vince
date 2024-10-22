package data

import (
	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/roaring"
)

var BitmapMarger = &pebble.Merger{
	Merge: func(key, value []byte) (pebble.ValueMerger, error) {
		ra := roaring.FromBufferWithCopy(value)
		return ra, nil
	},
	Name: "vince.BitmapMerger",
}
