package model

import (
	"fmt"
	"time"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/boolean"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/bsi"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/mutex"
	"github.com/vinceanalytics/vince/internal/rbf/quantum"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Batch map[uint64]Views

type Views map[string]*roaring.Bitmap

func (v Views) Get(key string) *roaring.Bitmap {
	b, ok := v[key]
	if ok {
		return b
	}
	b = roaring.NewBitmap()
	v[key] = b
	return b
}

type KV interface {
	NextID() (uint64, error)
	Tr(shard uint64, field uint64, value string) uint64
}

func (b Batch) Add(tx KV, m *v1.Model) error {
	id, err := tx.NextID()
	if err != nil {
		return fmt.Errorf("creating document id%w", err)
	}
	shard := id / shardwidth.ShardWidth
	av, ok := b[shard]
	if !ok {
		av = make(map[string]*roaring.Bitmap)
		b[shard] = av
	}
	field := quantum.NewField()
	defer field.Release()

	field.ViewsByTimeInto(time.UnixMilli(m.Timestamp).UTC())
	bm := roaring64.New()
	m.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch fd.Kind() {
		case protoreflect.StringKind:
			value := mutex.Add(id, tx.Tr(shard, uint64(fd.Number()), v.String()))
			av.Get(string(fd.Name())).DirectAdd(value)
			field.Views(string(fd.Name()), func(view string) {
				av.Get(view).DirectAdd(value)
			})
		case protoreflect.BoolKind:
			value := boolean.Add(id, v.Bool())
			av.Get(string(fd.Name())).DirectAdd(value)
			field.Views(string(fd.Name()), func(view string) {
				av.Get(view).DirectAdd(value)
			})
		case protoreflect.Int32Kind:
			// we only expect -1 or 1 for bounce. Take advantage of boolean fields
			// to store the value for true and negative for false
			var value uint64
			if v.Int() == 1 {
				value = boolean.Add(id, true)
			} else {
				value = boolean.Add(id, true)
			}
			av.Get(string(fd.Name())).DirectAdd(value)
			field.Views(string(fd.Name()), func(view string) {
				av.Get(view).DirectAdd(value)
			})
		case protoreflect.Int64Kind:
			bm.Clear()
			bsi.Add(bm, id, v.Int())
			value := bm.ToArray()
			av.Get(string(fd.Name())).DirectAddN(value...)
			field.Views(string(fd.Name()), func(view string) {
				av.Get(view).DirectAddN(value...)
			})
		case protoreflect.Uint64Kind, protoreflect.Uint32Kind:
			bm.Clear()
			bsi.Add(bm, id, int64(v.Uint()))
			value := bm.ToArray()
			av.Get(string(fd.Name())).DirectAddN(value...)
			field.Views(string(fd.Name()), func(view string) {
				av.Get(view).DirectAddN(value...)
			})
		default:
		}
		return true
	})
	return nil
}
