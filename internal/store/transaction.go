package store

import (
	"encoding/binary"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/web/query"
)

//go:generate protoc  --go_out=. --go_opt=paths=source_relative pages.proto

type Tx struct {
	tx    *badger.Txn
	it    *badger.Iterator
	enc   encoding.Encoding
	pos   int
	store *Store
}

var txPool = &sync.Pool{New: func() any {
	return &Tx{}
}}

func (o *Store) newTx(txn *badger.Txn) *Tx {
	tx := txPool.Get().(*Tx)
	tx.tx = txn
	tx.store = o
	return tx
}

func (tx *Tx) Iter() *badger.Iterator {
	if tx.it == nil {
		tx.it = tx.tx.NewIterator(badger.IteratorOptions{
			PrefetchSize:   32,
			PrefetchValues: true,
		})
	}
	return tx.it
}

func (tx *Tx) Close() {
	if tx.it != nil {
		tx.it.Close()
	}
	tx.it = nil
	tx.store = nil
}

func (tx *Tx) Release() {
	tx.Close()
	tx.tx = nil
	// avoid retaining large amout of data. Keep around 4kb per transaction. We
	// also use this to copy bitmaps
	tx.enc.Clip(4 << 10)
	tx.pos = 0
}

func (tx *Tx) Select(domain string, start,
	end time.Time, intrerval query.Interval, filters query.Filters, cb func(shard, view uint64, columns *roaring.Bitmap) error) error {
	m := tx.compile(filters)
	did := []byte(domain)
	return intrerval.Range(start, end, func(t time.Time) error {
		view := uint64(t.UnixMilli())
		for shard := range tx.Shards() {
			match := tx.Domain(shard, view, did)
			if match.IsEmpty() {
				return nil
			}
			columns := m.Apply(tx, shard, view, match)
			if columns.IsEmpty() {
				return nil
			}
			err := cb(shard, view, columns)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (tx *Tx) Shards() (n uint64) {
	it, err := tx.tx.Get(tx.enc.TranslateSeq(models.Field_unknown))
	if err == nil {
		it.Value(func(val []byte) error {
			n = binary.BigEndian.Uint64(val)
			return nil
		})
	}
	n = (n / ShardWidth) + 1
	return
}

func (tx *Tx) NewBitmap(shard, view uint64, field models.Field) (b *roaring.Bitmap) {
	key := encoding.Bitmap(view, shard, field, tx.enc.Allocate(encoding.BitmapKeySize))
	it, err := tx.tx.Get(key)
	if err == nil {
		it.Value(func(val []byte) error {
			dst := tx.enc.Allocate(len(val))
			copy(dst, val)
			b = roaring.FromBuffer(dst)
			return nil
		})
	}
	return b
}

func (tx *Tx) Domain(shard, view uint64, name []byte) *roaring.Bitmap {
	id := tx.store.ID(models.Field_domain, name)
	bs := tx.NewBitmap(shard, view, models.Field_domain)
	m := bs.Row(shard, id)
	return m
}

func (tx *Tx) Find(field models.Field, id uint64) (o string) {
	key := tx.enc.TranslateID(field, id)
	it, err := tx.tx.Get(key)
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			slog.Error("reading translation key", "key", err, "err", err)
		}
		return
	}
	it.Value(func(val []byte) error {
		o = string(val)
		return nil
	})
	return
}
