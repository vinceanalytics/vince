package ro2

import (
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Store struct {
	*DB
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
		id, err := tx.next()
		if err != nil {
			return err
		}
		shard := id / ro.ShardWidth
		re.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			b.Clear()
			switch fd.Kind() {
			case protoreflect.StringKind:
				ro.Mutex(b, id, tx.Tr(shard, uint64(fd.Number()), v.String()))
			case protoreflect.BoolKind:
				ro.Bool(b, id)
			case protoreflect.Int32Kind, protoreflect.Int64Kind:
				ro.BSI(b, id, v.Int())
			case protoreflect.Uint64Kind, protoreflect.Uint32Kind:
				ro.BSI(b, id, int64(v.Uint()))
			default:
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

func (o *Store) Seq() (id uint64) {
	o.View(func(tx *Tx) error {
		id = tx.Seq()
		return nil
	})
	return
}

func (o *Store) Shards(tx *Tx) (a []uint64) {
	q := tx.Seq()
	if q == 0 {
		return
	}
	e := q / ro.ShardWidth

	for i := uint64(0); i <= e; i++ {
		a = append(a, i)
	}
	return
}
