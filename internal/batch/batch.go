package batch

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/dgraph-io/badger/v4"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/encoding"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type KV interface {
	RecordID() uint64
	Translate(field v1.Field, value string) uint64
}

type Batch struct {
	data    map[encoding.Key]*roaring64.BSI
	domains map[string]uint64
}

func NewBatch() *Batch {
	return &Batch{
		data:    make(map[encoding.Key]*roaring64.BSI),
		domains: make(map[string]uint64),
	}
}

func (b *Batch) Add(tx KV, m *v1.Model) error {
	id := tx.RecordID()
	shard, ok := b.domains[m.Domain]
	if !ok {
		shard = tx.Translate(v1.Field_domain, m.Domain)
		b.domains[m.Domain] = shard
	}
	ts := uint64(time.UnixMilli(m.Timestamp).Truncate(time.Minute).UnixMilli())

	m.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		key := encoding.Key{
			Time:  ts,
			Shard: uint32(shard),
			Field: v1.Field(fd.Number()),
		}
		bs, ok := b.data[key]
		if !ok {
			bs = roaring64.NewDefaultBSI()
			b.data[key] = bs
		}
		var value int64
		switch fd.Kind() {
		case protoreflect.StringKind:
			value = int64(tx.Translate(key.Field, v.String()))
		case protoreflect.BoolKind:
			value = 1
		case protoreflect.Int32Kind, protoreflect.Int64Kind:
			value = v.Int()
		case protoreflect.Uint64Kind, protoreflect.Uint32Kind:
			value = int64(v.Uint())
		default:
			panic(fmt.Sprintf("unexpected model field kind%v", fd.Kind()))
		}
		bs.SetValue(id, value)
		return true
	})
	return nil
}

func (b *Batch) Save(db *badger.DB) (err error) {
	if len(b.data) == 0 {
		return
	}
	tx := db.NewTransaction(true)
	defer func() {
		if err != nil {
			tx.Discard()
		} else {
			err = tx.Commit()
		}
		clear(b.data)
	}()
	for k, v := range b.data {
		err = b.saveTs(tx, k, v)
		if err != nil {
			if errors.Is(err, badger.ErrTxnTooBig) {
				err = tx.Commit()
				if err != nil {
					return
				}
				tx = db.NewTransaction(true)
				err = b.saveTs(tx, k, v)
				if err != nil {
					return
				}
				continue
			}
			return err
		}
	}
	return nil
}

func (b *Batch) saveTs(tx *badger.Txn, key encoding.Key, value *roaring64.BSI) error {
	value.RunOptimize()
	ts := time.UnixMilli(int64(key.Time)).UTC()
	return errors.Join(
		b.saveKey(tx, encoding.Key{Field: key.Field}, value),                   // global
		b.saveKey(tx, encoding.Key{Field: key.Field, Shard: key.Shard}, value), // global by shard
		b.saveKey(tx, key, value),                                              // minute
		b.saveKey(tx, encoding.Key{Time: hour(ts), Shard: key.Shard, Field: key.Field}, value),
		b.saveKey(tx, encoding.Key{Time: day(ts), Shard: key.Shard, Field: key.Field}, value),
		b.saveKey(tx, encoding.Key{Time: week(ts), Shard: key.Shard, Field: key.Field}, value),
		b.saveKey(tx, encoding.Key{Time: month(ts), Shard: key.Shard, Field: key.Field}, value),
	)
}

func hour(ts time.Time) uint64 {
	return uint64(compute.Hour(ts).UnixMilli())
}

func day(ts time.Time) uint64 {
	return uint64(compute.Date(ts).UnixMilli())
}

func week(ts time.Time) uint64 {
	return uint64(compute.Week(ts).UnixMilli())
}

func month(ts time.Time) uint64 {
	return uint64(compute.Month(ts).UnixMilli())
}

func (b *Batch) saveKey(tx *badger.Txn, key encoding.Key, value *roaring64.BSI) error {
	return b.save(
		tx,
		encoding.EncodeKey(key),
		value,
	)
}

func (b *Batch) save(tx *badger.Txn, key []byte, value *roaring64.BSI) error {
	it, err := tx.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			b := get()
			defer put(b)
			_, err := value.WriteTo(b)
			if err != nil {
				return err
			}
			return tx.Set(key, bytes.Clone(b.Bytes()))
		}
		return err
	}
	exist := roaring64.NewDefaultBSI()
	err = it.Value(func(val []byte) error {
		_, err := exist.ReadFrom(bytes.NewReader(val))
		return err
	})
	if err != nil {
		return err
	}
	exist.ParOr(0, value)
	bs := get()
	defer put(bs)
	exist.RunOptimize()
	_, err = exist.WriteTo(bs)
	if err != nil {
		return err
	}
	return tx.Set(key, bytes.Clone(bs.Bytes()))
}

var bufferPool = &sync.Pool{New: func() any { return new(bytes.Buffer) }}

func get() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func put(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}
