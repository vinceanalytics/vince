package ro2

import (
	"hash/crc32"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Proto[T proto.Message] struct {
	*DB
	seq  atomic.Uint64
	ts   protoreflect.FieldDescriptor
	pool *sync.Pool
}

func Open[T proto.Message](path string) (*Proto[T], error) {
	var a T
	f := a.ProtoReflect().Descriptor().Fields()
	for i := 0; i < f.Len(); i++ {
		k := f.Get(i).Kind()
		assert(
			k == protoreflect.StringKind || k == protoreflect.Int64Kind,
			"unsupported field", "kind", k,
		)
	}
	ts := f.ByName(protoreflect.Name("timestamp"))
	assert(ts != nil, "timestamp field is required")
	db, err := New(path)
	if err != nil {
		return nil, err
	}
	o := &Proto[T]{DB: db, ts: ts, pool: &sync.Pool{}}
	o.pool.New = func() any {
		b := &bitmaps{
			pool:   o.pool,
			b:      make([]*roaring64.Bitmap, f.Len()),
			keys:   make([][]uint32, f.Len()),
			values: make([][]string, f.Len()),
		}
		for i := range b.b {
			b.b[i] = roaring64.New()
		}
		return b
	}
	return o, nil
}

func (o *Proto[T]) Add(msg T) error {
	return o.Update(func(tx *Tx) error {
		id := o.seq.Load()
		shard := id / ro.ShardWidth
		re := msg.ProtoReflect()
		ts := toDate(re.Get(o.ts).Int())
		b := o.get()
		defer b.release()
		hash := crc32.NewIEEE()
		re.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			if fd.Kind() == protoreflect.StringKind {
				hash.Reset()
				k := v.String()
				hash.Write([]byte(k))
				sum := hash.Sum32()
				b.get(fd).Add(
					ro.MutexPosition(id, uint64(sum)),
				)
				b.setKeys(fd, sum, k)
				return true
			}
			ro.BSI(b.get(fd), id, v.Int())
			return true
		})
		return b.each(func(field uint64, keys []uint32, values []string, bm *roaring64.Bitmap) error {
			return tx.Add(ts, shard, field, keys, values, bm)
		})
	})
}

func toDate(ts int64) uint64 {
	yy, mm, dd := time.UnixMilli(ts).UTC().Date()
	return uint64(time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC).UnixMilli())
}

func (o *Proto[T]) get() *bitmaps {
	return o.pool.Get().(*bitmaps)
}

type bitmaps struct {
	b      []*roaring64.Bitmap
	keys   [][]uint32
	values [][]string
	pool   *sync.Pool
}

func (b *bitmaps) get(f protoreflect.FieldDescriptor) *roaring64.Bitmap {
	return b.b[f.Number()-1]
}

func (b *bitmaps) setKeys(f protoreflect.FieldDescriptor, key uint32, value string) {
	idx := f.Number() - 1
	b.keys[idx] = append(b.keys[idx], key)
	b.values[idx] = append(b.values[idx], value)
}

func (b *bitmaps) each(f func(field uint64, keys []uint32, values []string, bm *roaring64.Bitmap) error) error {
	for i := range b.b {
		if b.b[i].IsEmpty() {
			continue
		}
		err := f(uint64(i+1), b.keys[i], b.values[i], b.b[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *bitmaps) release() {
	for i := range b.b {
		b.b[i].Clear()
		b.keys[i] = b.keys[i][:0]
		b.values[i] = b.values[i][:0]
	}
	b.pool.Put(b)
}
