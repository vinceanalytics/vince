package ro2

import (
	"bytes"
	"errors"
	"math"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

//go:generate protoc  --go_out=. --go_opt=paths=source_relative pages.proto

type Tx struct {
	tx *badger.Txn
	it *badger.Iterator
	// we need to retain keys until the transaction is commited
	keys []*alicia.Key
}

var txPool = &sync.Pool{New: func() any {
	return &Tx{}
}}

func newTx(txn *badger.Txn) *Tx {
	tx := txPool.Get().(*Tx)
	tx.tx = txn
	return tx
}

func (tx *Tx) get() *alicia.Key {
	k := alicia.Get()
	tx.keys = append(tx.keys, k)
	return k
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
	for i := range tx.keys {
		tx.keys[i].Release()
	}
	clear(tx.keys)
	tx.keys = tx.keys[:0]
	txPool.Put(tx)
}

func (tx *Tx) Depth(shard, field uint64) uint64 {
	if field == uint64(alicia.BOUNCE) {
		return 1
	}
	mx, ok := tx.max(shard, field)
	if !ok {
		return 0
	}
	return mx / ro.ShardWidth
}

func (tx *Tx) max(shard, field uint64) (uint64, bool) {
	prefix := tx.get().
		NS(alicia.CONTAINER).
		Shard(shard).
		Field(field).
		Key(math.MaxUint32).KeyPrefix()
	// seek to the last possible key
	it := tx.tx.NewIterator(badger.IteratorOptions{
		Reverse: true,
	})

	defer it.Close()
	it.Seek(prefix)
	if !it.Valid() {
		return 0, false
	}
	item := it.Item()
	key := alicia.Container(item.Key())
	var mx uint16
	item.Value(func(val []byte) error {
		mx = roaring.MaxOwned(item.UserMeta(), val)
		return nil
	})
	return (uint64(key) << 16) | uint64(mx), true
}

func (tx *Tx) Add(shard, field uint64, r *roaring64.Bitmap) error {
	return r.Save(&txWrite{
		shard: shard,
		field: field,
		tx:    tx,
	})
}

type txWrite struct {
	shard, field uint64
	tx           *Tx
}

var _ roaring64.Context = (*txWrite)(nil)

func (t *txWrite) Value(key uint32, cKey uint16, value func(uint8, []byte) error) error {
	xk := t.tx.get().
		NS(alicia.CONTAINER).
		Shard(t.shard).
		Field(t.field).
		Key(key).
		Container(cKey)
	it, err := t.tx.tx.Get(xk.Bytes())
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		return nil
	}
	return it.Value(func(val []byte) error {
		return value(it.UserMeta(), val)
	})
}

func (t *txWrite) Write(key uint32, cKey uint16, typ uint8, value []byte) error {
	xk := t.tx.get().
		NS(alicia.CONTAINER).
		Shard(t.shard).
		Field(t.field).
		Key(key).
		Container(cKey).Bytes()
	return t.tx.tx.SetEntry(badger.NewEntry(xk, value).WithMeta(typ))
}

func (tx *Tx) ExtractMutex(shard, field uint64, match *roaring64.Bitmap, f func(row uint64, value *roaring.Bitmap)) {
	filter := make([]*roaring.Container, 1<<ro.ShardVsContainerExponent)
	match.Each(func(key uint32, cKey uint16, value *roaring.Container) error {
		if value.IsEmpty() {
			return nil
		}
		idx := cKey % (1 << ro.ShardVsContainerExponent)
		filter[idx] = value
		return nil
	})
	prefix := tx.get().
		NS(alicia.CONTAINER).
		Shard(shard).
		Field(field).
		FieldPrefix()
	itr := tx.Iter()
	prevRow := ^uint64(0)
	seenThisRow := false
	var ac roaring.Container
	for itr.Seek(prefix); itr.ValidForPrefix(prefix); itr.Next() {
		item := itr.Item()
		k := alicia.Container(item.Key())
		row := uint64(k) >> ro.ShardVsContainerExponent
		if row == prevRow {
			if seenThisRow {
				continue
			}
		} else {
			seenThisRow = false
			prevRow = row
		}
		idx := k % (1 << ro.ShardVsContainerExponent)
		if filter[idx] == nil {
			continue
		}
		item.Value(func(val []byte) error {
			return ac.From(item.UserMeta(), val)
		})
		if ac.Intersects(filter[idx]) {
			f(row, ac.Bitmap())
		}
	}
}

func (tx *Tx) ExtractBool(shard, field uint64, match *roaring64.Bitmap, f func(row uint64, c int64)) {
	exists := tx.Row(shard, field, 0)
	exists.And(match)
	if exists.IsEmpty() {
		return
	}
	it := exists.Iterator()
	for it.HasNext() {
		f(it.Next(), 1)
	}
}

func (tx *Tx) ExtractBSI(shard, field uint64, match *roaring64.Bitmap, f func(row uint64, c int64)) {
	exists := tx.Row(shard, field, 0)
	if match != nil {
		exists.And(match)
	}
	if exists.IsEmpty() {
		return
	}
	data := make(map[uint64]uint64)
	mergeBits(exists, 0, data)

	sign := tx.Row(shard, field, 1)
	mergeBits(sign, 1<<63, data)

	bitDepth := tx.Depth(shard, field)

	for i := uint64(0); i < bitDepth; i++ {
		bits := tx.Row(shard, field, 2+uint64(i))
		if bits.IsEmpty() {
			continue
		}
		bits.And(exists)
		if bits.IsEmpty() {
			continue
		}
		mergeBits(bits, 1<<i, data)
	}
	for columnID, val := range data {
		val = uint64((2*(int64(val)>>63) + 1) * int64(val&^(1<<63)))
		f(columnID, int64(val))
	}
}

func mergeBits(bits *roaring64.Bitmap, mask uint64, out map[uint64]uint64) {
	it := bits.Iterator()
	for it.HasNext() {
		out[it.Next()] |= mask
	}
}

func (tx *Tx) Row(shard, field uint64, rowID uint64) *roaring64.Bitmap {
	o, from, to := tx.keyRange(shard, field, rowID)
	prefix := from.KeyPrefix()
	itr := tx.Iter()
	valid := func() bool {
		return itr.ValidForPrefix(prefix) &&
			bytes.Compare(itr.Item().Key(), to[:]) == -1
	}
	b := roaring.New()
	for itr.Seek(from[:]); valid(); itr.Next() {
		item := itr.Item()
		item.Value(func(val []byte) error {
			b.FromWire(0, item.UserMeta(), val)
			return nil
		})
	}
	o.Reset(b)
	return o
}

func (tx *Tx) keyRange(shard, field uint64, rowID uint64) (b *roaring64.Bitmap, from, to *alicia.Key) {
	b = roaring64.New()
	b.Add(rowID * ro.ShardWidth)
	b.Add((rowID + 1) * ro.ShardWidth)
	b.Each(func(key uint32, cKey uint16, value *roaring.Container) error {
		k := tx.get().NS(alicia.CONTAINER).Shard(shard).
			Field(field).Key(key).Container(cKey)
		if from == nil {
			from = k
			return nil
		}
		to = k
		return nil
	})
	return
}
