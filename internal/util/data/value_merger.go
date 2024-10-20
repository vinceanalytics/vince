package data

import (
	"bytes"
	"io"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/roaring"
)

type ValueMerger struct {
	ra *roaring.Bitmap
}

var _ pebble.ValueMerger = (*ValueMerger)(nil)

func (va *ValueMerger) MergeNewer(value []byte) error {
	va.ra.Or(roaring.FromBuffer(value))
	return nil
}

func (va *ValueMerger) MergeOlder(value []byte) error {
	va.ra.Or(roaring.FromBuffer(value))
	return nil
}

func (va *ValueMerger) Finish(includesBase bool) ([]byte, io.Closer, error) {
	return va.ra.ToBuffer(), nil, nil
}

var BitmapMarger = &pebble.Merger{
	Merge: func(key, value []byte) (pebble.ValueMerger, error) {
		ra := roaring.FromBuffer(bytes.Clone(value))
		return &ValueMerger{ra: ra}, nil
	},
	Name: "vince.BitmapMerger",
}
