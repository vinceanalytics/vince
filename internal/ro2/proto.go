package ro2

import (
	"hash"
	"hash/crc32"
	"sync"
	"sync/atomic"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Proto[T proto.Message] struct {
	*DB
	seq    atomic.Uint64
	ts     protoreflect.FieldDescriptor
	pool   *sync.Pool
	fields map[uint32]string
	names  map[string]uint32

	batch Batch
}

type Store = Proto[*v1.Model]

func Open(path string) (*Store, error) {
	return open[*v1.Model](path)
}

func open[T proto.Message](path string) (*Proto[T], error) {
	var a T
	f := a.ProtoReflect().Descriptor().Fields()
	fields := map[uint32]string{}
	names := map[string]uint32{}
	for i := 0; i < f.Len(); i++ {
		fd := f.Get(i)
		k := fd.Kind()
		assert(
			k == protoreflect.StringKind || k == protoreflect.Int64Kind,
			"unsupported field", "kind", k,
		)
		fields[uint32(fd.Number())] = string(fd.Name())
		names[string(fd.Name())] = uint32(fd.Number())
	}
	ts := f.ByName(protoreflect.Name("timestamp"))
	assert(ts != nil, "timestamp field is required")
	db, err := newDB(path)
	if err != nil {
		return nil, err
	}
	o := &Proto[T]{
		DB:     db,
		ts:     ts,
		fields: fields,
		names:  names,
		batch: Batch{
			hash: crc32.NewIEEE(),
		},
		pool: &sync.Pool{}}
	o.pool.New = func() any {
		b := &Bitmaps{
			pool:    o.pool,
			Roaring: make([]*roaring64.Bitmap, f.Len()),
			Keys:    make([][]uint32, f.Len()),
			Values:  make([][]string, f.Len()),
		}
		for i := range b.Roaring {
			b.Roaring[i] = roaring64.New()
		}
		return b
	}
	o.seq.Store(o.latestID(timestampField))
	return o, nil
}

func (o *Proto[T]) Name(number uint32) string {
	v, ok := o.fields[number]
	assert(ok, "unsupported field number", "number", number)
	return v
}

func (o *Proto[T]) Number(name string) uint32 {
	v, ok := o.names[name]
	assert(ok, "unsupported field name", "name", name)
	return v
}

func (o *Proto[T]) Flush() error {
	if o.batch.IsEmpty() {
		return nil
	}
	defer o.batch.Reset()

	return o.Update(func(tx *Tx) error {
		b := &o.batch
		for i := range b.shards {
			shard := b.shards[i]
			err := b.bitmaps[i].each(func(field uint64, keys []uint32, values []string, bm *roaring64.Bitmap) error {
				return tx.Add(shard, field, keys, values, bm)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (o *Proto[T]) Buffer(msg T) {
	id := o.seq.Add(1)
	shard := id / ro.ShardWidth
	re := msg.ProtoReflect()
	if len(o.batch.shards) == 0 || o.batch.shards[len(o.batch.shards)-1] != shard {
		// new batch
		o.batch.shards = append(o.batch.shards, shard)
		o.batch.bitmaps = append(o.batch.bitmaps, o.get())
	}
	b := o.batch.bitmaps[len(o.batch.bitmaps)-1]
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

}

func (o *Proto[T]) Batch(msgs []T) *Batch {
	s := &Batch{}
	curr := ^uint64(0)
	b := o.get()
	hash := crc32.NewIEEE()
	for i := range msgs {
		id := o.seq.Add(1)
		shard := id / ro.ShardWidth
		if shard != curr {
			if i != 0 {
				s.shards = append(s.shards, curr)
				s.bitmaps = append(s.bitmaps, b)
				b = o.get()
			}
			curr = shard
		}
		msg := msgs[i]
		re := msg.ProtoReflect()
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

	}
	return s
}

func (o *Proto[T]) get() *Bitmaps {
	return o.pool.Get().(*Bitmaps)
}

type Batch struct {
	shards  []uint64
	bitmaps []*Bitmaps
	hash    hash.Hash32
}

func (s *Batch) IsEmpty() bool {
	return len(s.shards) == 0
}

func (s *Batch) Reset() {
	for i := range s.bitmaps {
		s.bitmaps[i].release()
	}
	clear(s.bitmaps)
	s.bitmaps = s.bitmaps[:0]
	s.shards = s.shards[:0]
}

// Bitmaps index of all fields
type Bitmaps struct {
	// 2d transformed data to bitmaps
	Roaring []*roaring64.Bitmap
	// crc32 hash of the values for string keys.
	Keys [][]uint32
	// String keys
	Values [][]string
	pool   *sync.Pool
}

func (b *Bitmaps) get(f protoreflect.FieldDescriptor) *roaring64.Bitmap {
	return b.Roaring[f.Number()-1]
}

func (b *Bitmaps) setKeys(f protoreflect.FieldDescriptor, key uint32, value string) {
	idx := f.Number() - 1
	b.Keys[idx] = append(b.Keys[idx], key)
	b.Values[idx] = append(b.Values[idx], value)
}

func (b *Bitmaps) each(f func(field uint64, keys []uint32, values []string, bm *roaring64.Bitmap) error) error {
	for i := range b.Roaring {
		if b.Roaring[i].IsEmpty() {
			continue
		}
		err := f(uint64(i+1), b.Keys[i], b.Values[i], b.Roaring[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bitmaps) release() {
	for i := range b.Roaring {
		b.Roaring[i].Clear()
		b.Keys[i] = b.Keys[i][:0]
		b.Values[i] = b.Values[i][:0]
	}
	b.pool.Put(b)
}
