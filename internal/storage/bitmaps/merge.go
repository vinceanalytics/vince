package bitmaps

import (
	"io"

	"github.com/cockroachdb/pebble"
	"github.com/gernest/roaring"
)

type Bitmap = roaring.Bitmap

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
	return co.Encode(), nil, nil
}

func (m *Merger) merge(value []byte) error {
	n := roaring.Union((*roaring.Container)(m), roaring.DecodeContainer(value))
	// We take advantage of the fact that we don't store random numbers which makes it
	// favorable for run containers.
	//
	// This signficatnly leads to be best storage savings.
	co := (*Merger)(n.Optimize())
	*m = *co
	return nil
}

var Merge = &pebble.Merger{
	Name: "roaring.Container",
	Merge: func(key, value []byte) (pebble.ValueMerger, error) {
		co := roaring.DecodeContainer(value).Clone()
		return (*Merger)(co), nil
	},
}
