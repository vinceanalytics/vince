package ro2

import (
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
	fields map[uint32]string
	names  map[string]uint32
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

func (o *Proto[T]) One(msg T) error {
	return o.Update(func(tx *Tx) error {
		re := msg.ProtoReflect()
		b := roaring64.New()
		var err error
		id := o.seq.Add(1)
		shard := id / ro.ShardWidth

		re.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			b.Clear()
			if fd.Kind() == protoreflect.StringKind {
				b.Add(
					ro.MutexPosition(id, tx.Tr(shard, uint64(fd.Number()), v.String())),
				)
			} else {
				ro.BSI(b, id, v.Int())
			}
			err = tx.Add(shard, uint64(fd.Number()), b)
			return err == nil
		})
		return err
	})
}
