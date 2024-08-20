package ro2

import (
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

	buffer []T
	tr     translator
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
		buffer: make([]T, 0, 4<<10),
		pool:   &sync.Pool{}}
	o.pool.New = func() any {
		b := &Bitmaps{
			pool:    o.pool,
			Roaring: make([]*roaring64.Bitmap, f.Len()),
		}
		for i := range b.Roaring {
			b.Roaring[i] = roaring64.New()
		}
		return b
	}
	o.View(func(tx *Tx) error {
		o.tr.init(tx)
		return nil
	})
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
	return o.apply()
}

func (o *Proto[T]) Buffer(msg T) {
	o.buffer = append(o.buffer, proto.Clone(msg).(T))
}

func (o *Proto[T]) apply() error {
	if len(o.buffer) == 0 {
		return nil
	}
	defer func() {
		clear(o.buffer)
		o.buffer = o.buffer[:0]
		o.tr.tx = nil
	}()
	return o.Update(func(tx *Tx) error {
		b := o.get()
		defer b.release()

		o.tr.Reset(tx)

		curr := ^uint64(0)
		for i := range o.buffer {
			id := o.seq.Add(1)
			shard := id / ro.ShardWidth
			if shard != curr {
				if i != 0 {
					err := b.each(func(field uint64, bm *roaring64.Bitmap) error {
						return tx.Add(shard, field, bm)
					})
					if err != nil {
						return err
					}
				}
				curr = shard
				b.reset()
			}
			msg := o.buffer[i]
			re := msg.ProtoReflect()
			re.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
				if fd.Kind() == protoreflect.StringKind {
					b.get(fd).Add(
						ro.MutexPosition(id, o.tr.Tr(shard, uint64(fd.Number()), v.String())),
					)
					return true
				}
				ro.BSI(b.get(fd), id, v.Int())
				return true
			})
		}
		return b.each(func(field uint64, bm *roaring64.Bitmap) error {
			return tx.Add(curr, field, bm)
		})
	})
}

func (o *Proto[T]) get() *Bitmaps {
	return o.pool.Get().(*Bitmaps)
}

// Bitmaps index of all fields
type Bitmaps struct {
	// 2d transformed data to bitmaps
	Roaring []*roaring64.Bitmap
	pool    *sync.Pool
}

func (b *Bitmaps) get(f protoreflect.FieldDescriptor) *roaring64.Bitmap {
	return b.Roaring[f.Number()-1]
}

func (b *Bitmaps) each(f func(field uint64, bm *roaring64.Bitmap) error) error {
	for i := range b.Roaring {
		if b.Roaring[i].IsEmpty() {
			continue
		}
		err := f(uint64(i+1), b.Roaring[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bitmaps) release() {
	b.reset()
	b.pool.Put(b)
}

func (b *Bitmaps) reset() {
	for i := range b.Roaring {
		b.Roaring[i].Clear()
	}
}
