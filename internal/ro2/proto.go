package ro2

import (
	"math"
	"sync/atomic"

	"github.com/dgraph-io/badger/v4"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Store struct {
	*DB
	seq atomic.Uint64
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
		it := tx.tx.NewIterator(badger.IteratorOptions{
			Reverse: true,
		})
		defer it.Close()
		prefix := tx.get().
			NS(alicia.CONTAINER).
			Shard(math.MaxUint32).Field(uint64(alicia.TIMESTAMP)).ShardPrefix()
		it.Seek(prefix)
		if !it.Valid() {
			return nil
		}
		shard := alicia.Shard(it.Item().Key())

		// we choose the last shard to look for the last saved sequence ID. We
		// already know that timestamp field is required for all events, since
		// bsi fields store existence bitmap as row 0, we can safely derive tha
		// last sequence ID as highest existence bit set.
		exists := tx.Row(uint64(shard), uint64(alicia.TIMESTAMP), 0)
		if !exists.IsEmpty() {
			o.seq.Store(exists.Maximum())
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
		re.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			b.Clear()
			if fd.Kind() == protoreflect.StringKind {
				ro.BSI(b, id, int64(tx.Tr(shard, uint64(fd.Number()), v.String())))
			} else {
				ro.BSI(b, id, v.Int())
			}
			err = tx.Add(shard, uint64(fd.Number()), b)
			return err == nil
		})
		if err != nil {
			return err
		}
		ts := re.Get(tsField).Int()
		return tx.quatum(b, shard, id, ts)
	})
}

func (o *Store) shards() (a []uint64) {
	q := o.seq.Load()
	if q == 0 {
		return
	}
	e := q / ro.ShardWidth

	for i := uint64(0); i <= e; i++ {
		a = append(a, i)
	}
	return
}
