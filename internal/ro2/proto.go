package ro2

import (
	"math"
	"sync/atomic"

	"github.com/dgraph-io/badger/v4"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
	"github.com/vinceanalytics/vince/internal/shards"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Store struct {
	*DB
	seq    atomic.Uint64
	shards shards.Shards
}

func Open(path string) (*Store, error) {
	return open(path)
}

func open(path string) (*Store, error) {
	db, err := newDB(path)
	if err != nil {
		return nil, err
	}
	o := &Store{
		DB: db,
	}
	err = o.View(func(tx *Tx) error {
		// try loading shards / ts mapping
		if it, err := tx.tx.Get(tx.get().Shards()); err == nil {
			var vs v1.Shards
			it.Value(func(val []byte) error {
				return proto.Unmarshal(val, &vs)
			})
			o.shards.Set(vs.Shards, vs.Ts)
		}

		// load last sequence from timestamp existence field
		key := tx.get().
			NS(alicia.CONTAINER).
			Field(uint64(alicia.TIMESTAMP)).
			Shard(uint64(math.MaxUint32))

		it := tx.tx.NewIterator(badger.IteratorOptions{
			Reverse: true,
		})
		defer it.Close()
		it.Seek(key.ShardPrefix())
		if it.Valid() {
			shard := alicia.Shard(it.Item().Key())
			exists := tx.Row(uint64(shard), uint64(alicia.TIMESTAMP), 0)
			if !exists.IsEmpty() {
				o.seq.Store(exists.Maximum())
				return nil
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return o, nil
}

var (
	fields  = new(v1.Model).ProtoReflect().Descriptor().Fields()
	tsField = fields.ByNumber(protowire.Number(alicia.TIMESTAMP))
)

func (o *Store) Name(number uint32) string {
	f := fields.ByNumber(protowire.Number(number))
	return string(f.Name())
}

func (o *Store) Number(name string) uint32 {
	f := fields.ByName(protoreflect.Name(name))
	return uint32(f.Number())
}

func (o *Store) One(msg *v1.Model) error {
	return o.Update(func(tx *Tx) error {
		re := msg.ProtoReflect()
		b := roaring64.New()
		var err error
		id := o.seq.Add(1)
		shard := id / ro.ShardWidth
		ts := re.Get(tsField)
		if o.shards.Add(shard, ts.Int()) {
			// change  in shards. Persist shard/ts state
			vs := &v1.Shards{}
			vs.Shards, vs.Ts = o.shards.Load()
			data, _ := proto.Marshal(vs)
			err := tx.tx.Set(tx.get().Shards(), data)
			if err != nil {
				return err
			}
		}
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
