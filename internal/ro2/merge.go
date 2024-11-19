package ro2

import (
	"io"

	"github.com/cockroachdb/pebble"
	"github.com/gernest/roaring"
)

type Bitmap = roaring.Bitmap

var NewBitmap = roaring.NewBitmap

type Merger roaring.Container

var _ pebble.ValueMerger = (*Merger)(nil)

func (m *Merger) MergeNewer(value []byte) error {
	return m.merge(value)
}

func (m *Merger) MergeOlder(value []byte) error {
	return m.merge(value)
}

func (m *Merger) Finish(includesBase bool) ([]byte, io.Closer, error) {
	co := (*roaring.Container)(m)
	co.Optimize()
	return EncodeContainer(co), nil, nil
}

func (m *Merger) merge(value []byte) error {
	n := roaring.Union((*roaring.Container)(m), DecodeContainer(value))
	co := (*Merger)(n)
	*m = *co
	return nil
}

var Merge = &pebble.Merger{
	Name: "pilosa.RoaringBitmap",
	Merge: func(key, value []byte) (pebble.ValueMerger, error) {
		co := DecodeContainer(value).Clone()
		return (*Merger)(co), nil
	},
}
