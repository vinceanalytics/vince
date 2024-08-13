package oracle

import (
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/assert"
	"github.com/vinceanalytics/vince/internal/btx"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	seq = []byte("seq")
)

const (
	zero   = ^uint64(0)
	ID     = "_id"
	layout = "2006010215"
)

const (
	// BSI bits used to check existence & sign.
	bsiExistsBit = 0
	bsiSignBit   = 1
	bsiOffsetBit = 2
)

func (d *dbShard) Write(e *v1.Model) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		s, err := tx.CreateBucketIfNotExists(seq)
		if err != nil {
			return err
		}
		id, err := s.NextSequence()
		if err != nil {
			return err
		}
		shard := id / shardwidth.ShardWidth
		b := make(bitmaps)
		b.get(ID).DirectAdd(id % shardwidth.ShardWidth)

		wInt := func(name string, v int64) {
			btx.BSI(b.get(name), id, v)
		}
		wString := func(name string, v string) {
			f, err := newWriteField(tx, []byte(name))
			assert.Assert(err == nil, "create new write field", "field", name, "err", err)
			key, err := f.translate([]byte(v))
			assert.Assert(err == nil, "translate value", "field", name, "value", v, "err", err)
			btx.Mutex(b.get(name), id, key)
		}
		e.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			name := string(fd.Name())
			switch fd.Kind() {
			case protoreflect.StringKind:
				wString(name, v.String())
			case protoreflect.BoolKind:
				wInt(name, 1)
			case protoreflect.Int64Kind:
				if name == "timestamp" {
					ts := v.Int()
					wInt(name, ts)
					wString("date", date(ts))
				} else {
					wInt(name, v.Int())
				}

			case protoreflect.Uint64Kind, protoreflect.Uint32Kind:
				wInt(name, int64(v.Uint()))
			default:
				if fd.IsMap() {
					prefix := name + "."
					v.Map().Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
						wString(prefix+mk.String(), v.String())
						return true
					})
				}
			}
			return true
		})
		// For bounce, session, view and duration fields we only perform sum on the
		// bsi. To save space we don't bother storing zero values.
		//
		// null bounce act as a clear signal , we set it to -1 so that when
		// a user stay on the site and navigated to a different page during
		// a live session the result will be 0 [first page sets 1, next page sets -1
		// and subsequent pages within the same session will be 0].
		if e.Bounce == nil {
			wInt("bounce", -1)
		}
		return d.saveBitmaps(shard, e.Timestamp, b)
	})
}

func (d *dbShard) saveBitmaps(shard uint64, ts int64, b bitmaps) error {
	db, err := d.open(shard)
	if err != nil {
		return err
	}
	tx, err := db.db.Begin(true)
	if err != nil {
		return err
	}
	for k, v := range b {
		_, err = tx.AddRoaring(k, v)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	db.updateTS(ts)
	return nil
}

type bitmaps map[string]*roaring.Bitmap

func (b bitmaps) get(name string) *roaring.Bitmap {
	o, ok := b[name]
	if !ok {
		o = roaring.NewBitmap()
		b[name] = o
	}
	return o
}
