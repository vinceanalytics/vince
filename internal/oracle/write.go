package oracle

import (
	"math/bits"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"go.etcd.io/bbolt"
)

var (
	seq = []byte("seq")
)

const (
	zero   = ^uint64(0)
	ID     = "_id"
	layout = "2006010215"
)

const (
	// BSI bits used to check existence & sign.
	bsiExistsBit = 0
	bsiSignBit   = 1
	bsiOffsetBit = 2
)

type Writer interface {
	Write(ts int64, f func(columns Columns) error) error
}

type Columns interface {
	String(field string, value string) error
	Int64(field string, value int64) error
}

type write struct {
	seq      *bbolt.Bucket
	data     map[string]*roaring.Bitmap
	fields   map[string]*field
	tx       *bbolt.Tx
	id       uint64
	shard    uint64
	min, max int64
}

func (w *write) Write(ts int64, f func(columns Columns) error) error {
	id, err := w.seq.NextSequence()
	if err != nil {
		return err
	}
	shard := id / shardwidth.ShardWidth
	if shard != w.shard {
		if w.shard != shard {
			err := w.flush()
			if err != nil {
				return err
			}
		}
		w.shard = shard
		w.min = ts
		w.max = ts
	}
	w.id = id
	w.min = min(w.min, ts)
	w.max = max(w.max, ts)
	return f(w)
}

func (w *write) Int64(field string, svalue int64) error {
	m := w.get(field)
	id := w.id
	fragmentColumn := id % shardwidth.ShardWidth
	m.DirectAdd(fragmentColumn)
	negative := svalue < 0
	var value uint64
	if negative {
		m.Add(shardwidth.ShardWidth + fragmentColumn) // set sign bit
		value = uint64(svalue * -1)
	} else {
		value = uint64(svalue)
	}
	lz := bits.LeadingZeros64(value)
	row := uint64(2)
	for mask := uint64(0x1); mask <= 1<<(64-lz) && mask != 0; mask = mask << 1 {
		if value&mask > 0 {
			m.DirectAdd(row*shardwidth.ShardWidth + fragmentColumn)
		}
		row++
	}
	return nil
}

func (w *write) String(field string, value string) error {
	m, f, err := w.getString(field)
	if err != nil {
		return err
	}
	v, err := f.translate([]byte(value))
	if err != nil {
		return err
	}
	m.DirectAdd(v*shardwidth.ShardWidth + (w.id % shardwidth.ShardWidth))
	return nil
}

func (w *write) getString(name string) (*roaring.Bitmap, *field, error) {
	if b, ok := w.data["name"]; ok {
		return b, w.fields[name], nil
	}
	b := roaring.NewBitmap()
	f, err := newWriteField(w.tx, []byte(name))
	if err != nil {
		return nil, nil, err
	}
	w.data[name] = b
	w.fields[name] = f
	return b, f, nil
}

func (w *write) get(name string) *roaring.Bitmap {
	b, ok := w.data[name]
	if !ok {
		b = roaring.NewBitmap()
		w.data[name] = b
	}
	return b
}

func (w *write) flush() error {
	return nil
}
