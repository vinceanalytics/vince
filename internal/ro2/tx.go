package ro2

import (
	"bytes"
	"errors"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/dgraph-io/badger/v4"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/web/query"
)

//go:generate protoc  --go_out=. --go_opt=paths=source_relative pages.proto

type Tx struct {
	tx *badger.Txn
	it *badger.Iterator
}

var txPool = &sync.Pool{New: func() any {
	return &Tx{}
}}

func newTx(txn *badger.Txn) *Tx {
	tx := txPool.Get().(*Tx)
	tx.tx = txn
	return tx
}

func (tx *Tx) Iter() *badger.Iterator {
	if tx.it == nil {
		tx.it = tx.tx.NewIterator(badger.IteratorOptions{})
	}
	return tx.it
}

func (tx *Tx) Close() {
	if tx.it != nil {
		tx.it.Close()
	}
	tx.it = nil
}

func (tx *Tx) Release() {
	tx.Close()
	tx.tx = nil
}

func (tx *Tx) Select(domain string, start,
	end time.Time, intrerval query.Interval, filters query.Filters, cb func(shard, view uint64, columns *roaring64.Bitmap) error) error {
	shard, ok := tx.ID(v1.Field_domain, domain)
	if !ok {
		return nil
	}
	m := tx.compile(filters)
	return intrerval.Range(start, end, func(t time.Time) error {
		view := uint64(t.UnixMilli())
		bs, err := tx.Bitmap(shard, view, v1.Field_domain)
		if err != nil {
			return err
		}
		match := bs.GetExistenceBitmap()
		columns := m.Apply(tx, shard, view, match)
		if columns.IsEmpty() {
			return nil
		}
		return cb(shard, view, match)
	})
}

func (tx *Tx) Sum(shard, view uint64, field v1.Field, match *roaring64.Bitmap) (int64, error) {
	bs, err := tx.Bitmap(shard, view, field)
	if err != nil {
		return 0, err
	}
	sum, _ := bs.Sum(match)
	return sum, nil
}

func (tx *Tx) Unique(shard, view uint64, field v1.Field, match *roaring64.Bitmap) (uint64, error) {
	bs, err := tx.Bitmap(shard, view, field)
	if err != nil {
		return 0, err
	}
	return bs.IntersectAndTranspose(0, match).GetCardinality(), nil
}

func (tx *Tx) Count(shard, view uint64, field v1.Field, match *roaring64.Bitmap) (uint64, error) {
	bs, err := tx.Bitmap(shard, view, field)
	if err != nil {
		return 0, err
	}
	set := roaring64.And(bs.GetExistenceBitmap(), match)
	return set.GetCardinality(), nil
}

func (tx *Tx) Bitmap(shard, view uint64, field v1.Field) (*roaring64.BSI, error) {
	key := encoding.EncodeKey(encoding.Key{
		Time:  view,
		Shard: uint32(shard),
		Field: field,
	})
	it, err := tx.tx.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return roaring64.NewDefaultBSI(), nil
		}
		return nil, err
	}
	bs := roaring64.NewDefaultBSI()
	return bs, it.Value(func(val []byte) error {
		_, err := bs.ReadFrom(bytes.NewReader(val))
		return err
	})
}
