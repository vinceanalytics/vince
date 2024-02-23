package events

import (
	"fmt"
	"sort"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/memory"
	v1 "github.com/vinceanalytics/vince/gen/go/events/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var Mapping, Schema = mapping()

func mapping() (map[string]int, *arrow.Schema) {
	b := New(memory.NewGoAllocator())
	defer b.Release()
	r := b.NewRecord()
	defer r.Release()
	o := make(map[string]int)
	for i := 0; i < int(r.NumCols()); i++ {
		o[r.ColumnName(i)] = i
	}
	return o, r.Schema()
}

type Builder struct {
	r      *array.RecordBuilder
	fields map[protoreflect.FieldNumber]buildFunc
}

type field struct {
	number protoreflect.FieldNumber
	arrow  arrow.Field
}

func New(mem memory.Allocator) *Builder {
	dt := &v1.Data{}
	fs := dt.ProtoReflect()
	fds := fs.Descriptor().Fields()
	fields := make([]*field, 0, fds.Len())
	for i := 0; i < fds.Len(); i++ {
		f := fds.Get(i)
		fields = append(fields, &field{
			number: f.Number(),
			arrow: arrow.Field{
				Name:     string(f.Name()),
				Nullable: f.HasOptionalKeyword() || f.Kind() == protoreflect.StringKind,
				Type:     kinds[f.Kind()],
			},
		})
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].number < fields[j].number
	})
	af := make([]arrow.Field, 0, len(fields))
	for i := range fields {
		af = append(af, fields[i].arrow)
	}
	r := array.NewRecordBuilder(mem, arrow.NewSchema(af, nil))
	fm := make(map[protoreflect.FieldNumber]buildFunc)
	for i, f := range r.Fields() {
		fm[fields[i].number] = newBuild(f)
	}
	return &Builder{r: r, fields: fm}
}

func (b *Builder) NewRecord() arrow.Record {
	return b.r.NewRecord()
}

func (b *Builder) Release() {
	b.r.Release()
}

func (b *Builder) Write(list *v1.List) arrow.Record {
	ls := list.GetItems()
	sort.SliceStable(ls, func(i, j int) bool {
		return ls[i].Timestamp < ls[j].Timestamp
	})
	b.r.Reserve(len(ls))
	for _, e := range ls {
		b.writeData(e)
	}
	return b.NewRecord()
}

func (b *Builder) writeData(data *v1.Data) {
	fs := data.ProtoReflect()
	fds := fs.Descriptor().Fields()
	for i := 0; i < fds.Len(); i++ {
		f := fds.Get(i)
		if f.HasOptionalKeyword() && !fs.Has(f) {
			b.fields[f.Number()](protoreflect.ValueOf(nil))
			continue
		}
		b.fields[f.Number()](fs.Get(f))
	}
}

var kinds = map[protoreflect.Kind]arrow.DataType{
	protoreflect.Int64Kind:  arrow.PrimitiveTypes.Int64,
	protoreflect.DoubleKind: arrow.PrimitiveTypes.Float64,
	protoreflect.BoolKind:   arrow.FixedWidthTypes.Boolean,
	protoreflect.StringKind: &arrow.DictionaryType{
		IndexType: arrow.PrimitiveTypes.Uint32,
		ValueType: arrow.BinaryTypes.String,
	},
}

func newBuild(a array.Builder) buildFunc {
	switch e := a.(type) {
	case *array.Int64Builder:
		return func(fv protoreflect.Value) {
			e.Append(fv.Int())
		}
	case *array.Float64Builder:
		return func(fv protoreflect.Value) {
			e.Append(fv.Float())
		}
	case *array.BooleanBuilder:
		return func(fv protoreflect.Value) {
			if !fv.IsValid() {
				e.AppendNull()
				return
			}
			e.Append(fv.Bool())
		}
	case *array.BinaryDictionaryBuilder:
		return func(fv protoreflect.Value) {
			v := fv.String()
			if v == "" {
				e.AppendNull()
				return
			}
			e.Append([]byte(v))
		}
	default:
		panic(fmt.Sprintf("%T is not supported", e))
	}
}

type buildFunc func(fv protoreflect.Value)
