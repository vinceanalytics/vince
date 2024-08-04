package oracle

import (
	"fmt"

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
	Int64(field string, value int64)
	Bool(field string, value bool)
}

func (d *dbShard) Write() (*write, error) {
	tx, err := d.db.Begin(true)
	if err != nil {
		return nil, err
	}
	s, err := tx.CreateBucketIfNotExists(seq)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	w := &write{
		tx:     tx,
		seq:    s,
		data:   make(map[string]*roaring.Bitmap),
		fields: make(map[string]*field),
		db:     d,
	}
	return w, nil
}

type write struct {
	seq      *bbolt.Bucket
	data     map[string]*roaring.Bitmap
	fields   map[string]*field
	tx       *bbolt.Tx
	id       uint64
	shard    uint64
	min, max int64
	db       *dbShard
}

func (w *write) Close() error {
	defer func() {
		w.tx.Commit()
	}()
	return w.flush()
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
	// set existence column
	w.get(ID).DirectAdd(w.id % shardwidth.ShardWidth)
	return f(w)
}

func (w *write) Bool(field string, value bool) {
	if value {
		w.Int64(field, 1)
		return
	}
	w.Int64(field, 0)
}

func (w *write) Int64(field string, svalue int64) {
	setValue(w.get(field), w.id, svalue)
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
	defer func() {
		clear(w.data)
		w.min = 0
		w.max = 0
	}()
	db, err := w.db.open(w.shard)
	if err != nil {
		return fmt.Errorf("open shard db %w", err)
	}
	tx, err := db.db.Begin(true)
	if err != nil {
		return err
	}
	for k, d := range w.data {
		_, err := tx.AddRoaring(k, d)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("adding data %w", err)
		}
	}
	lo, hi := db.min.Load(), db.max.Load()
	if lo == 0 {
		lo = w.min
	}
	db.min.Store(min(lo, w.min))
	db.max.Store(max(hi, w.max))
	return tx.Commit()
}
