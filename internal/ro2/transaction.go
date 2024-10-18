package ro2

import (
	"encoding/binary"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/y"
	"github.com/vinceanalytics/vince/internal/bsi"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/web/query"
)

//go:generate protoc  --go_out=. --go_opt=paths=source_relative pages.proto

type Tx struct {
	tx      *badger.Txn
	it      *badger.Iterator
	enc     encoding.Encoding
	bsi     map[uint32]*bsi.BSI
	bitmaps []*roaring.Bitmap
	kv      [31]bsi.BSI
	pos     int
	store   *Store
}

var txPool = &sync.Pool{New: func() any {
	return &Tx{bsi: make(map[uint32]*bsi.BSI)}
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
	clear(tx.bsi)
	clear(tx.bitmaps)
	tx.bitmaps = tx.bitmaps[:0]
	clear(tx.kv[:])
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

func (tx *Tx) Sum(shard, view uint64, field models.Field, match *roaring.Bitmap) (sum int64) {
	bs := tx.Bitmap(shard, view, field)
	sum, _ = bs.Sum(match)
	return
}

func (tx *Tx) Transpose(shard, view uint64, field models.Field, match *roaring.Bitmap) (result *roaring.Bitmap) {
	bs := tx.Bitmap(shard, view, field)
	return bs.Transpose(match)
}

func (tx *Tx) TransposeSet(shard, view uint64, field models.Field, match *roaring.Bitmap) (result map[int64][]uint64) {
	bs := tx.Bitmap(shard, view, field)
	tr := bs.Extract(match)
	result = make(map[int64][]uint64)
	for k, v := range tr {
		result[v] = append(result[v], k)
	}
	return
}

func (tx *Tx) Count(shard, view uint64, field models.Field, match *roaring.Bitmap) (count uint64) {
	bs := tx.Bitmap(shard, view, field)
	ex := bs.GetExistenceBitmap()
	if ex != nil {
		count = ex.AndCardinality(match)
	}
	return
}

func (tx *Tx) Bitmap(shard, view uint64, field models.Field) *bsi.BSI {
	key := encoding.Bitmap(view, shard, field, 0, tx.enc.Allocate(encoding.BitmapKeySize))
	prefix := key[:len(key)-1]
	kh := y.Hash(prefix)
	if b, ok := tx.bsi[kh]; ok {
		return b
	}
	b := tx.newKv(key)
	tx.bsi[kh] = b
	return b
}

func (tx *Tx) Domain(shard, view uint64, name []byte) *roaring.Bitmap {
	id := tx.store.ID(models.Field_domain, name)
	return tx.Compare(
		shard, view, models.Field_domain, bsi.EQ, int64(id), 0, nil,
	)
}

func (tx *Tx) Compare(shard, view uint64, field models.Field,
	op bsi.Operation, valueOrStart, end int64, foundSet *roaring.Bitmap) *roaring.Bitmap {
	bs := tx.Bitmap(shard, view, field)
	return bs.CompareValue(0, op, valueOrStart, end, foundSet)
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
