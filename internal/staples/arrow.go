package staples

import (
	"fmt"
	"reflect"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/vinceanalytics/vince/internal/camel"
	"github.com/vinceanalytics/vince/internal/cluster/events"
)

type Arrow[T any] struct {
	build  *array.RecordBuilder
	append func(reflect.Value)
}

func (a *Arrow[T]) NewRecord() arrow.Record {
	return a.build.NewRecord()
}
func (a *Arrow[T]) Release() {
	a.build.Release()
}

func NewArrow[T any](mem memory.Allocator) *Arrow[T] {
	b, w := NewRecordBuilder(mem, Schema[T]())
	return &Arrow[T]{
		build:  b,
		append: w,
	}
}

func Schema[T any]() *arrow.Schema {
	var a T
	return schemaOf(a)
}

func (a *Arrow[T]) Append(v *T) {
	a.append(reflect.ValueOf(v))
}

func schemaOf(a any) *arrow.Schema {
	return arrow.NewSchema(
		build(reflect.TypeOf(a)),
		nil,
	)
}

func build(r reflect.Type) (o []arrow.Field) {
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}
	if r.Kind() != reflect.Struct {
		panic("only structs are supported")
	}
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)
		typ := f.Type
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		if base, ok := baseTypes[typ.Kind()]; ok {
			o = append(o, arrow.Field{
				Name:     camel.Case(f.Name),
				Type:     base,
				Nullable: f.Type.Kind() == reflect.Ptr || typ.Kind() == reflect.String,
			})
			continue
		}
		panic(typ.String() + " slices are not supported")
	}
	return
}

var baseTypes = map[reflect.Kind]arrow.DataType{
	reflect.Bool: arrow.FixedWidthTypes.Boolean,
	reflect.String: &arrow.DictionaryType{
		IndexType: arrow.PrimitiveTypes.Uint32,
		ValueType: arrow.BinaryTypes.String,
	},
	reflect.Int64:   arrow.PrimitiveTypes.Int64,
	reflect.Float64: arrow.PrimitiveTypes.Float64,
}

func NewRecordBuilder(mem memory.Allocator, as *arrow.Schema) (*array.RecordBuilder, func(reflect.Value)) {
	b := array.NewRecordBuilder(mem, as)
	fields := make([]func(reflect.Value), len(b.Fields()))
	for i := range fields {
		fields[i] = write(b.Field(i))
	}
	return b, func(v reflect.Value) {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		for i := 0; i < v.NumField(); i++ {
			fields[i](v.Field(i))
		}
	}
}

func write(b array.Builder) func(reflect.Value) {
	switch e := b.(type) {
	case *array.BooleanBuilder:
		return func(v reflect.Value) {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					e.AppendNull()
					return
				}
				v = v.Elem()
			}
			e.Append(v.Bool())
		}
	case *array.Int64Builder:
		return func(v reflect.Value) {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					e.AppendNull()
					return
				}
				v = v.Elem()
			}
			e.Append(v.Int())
		}
	case *array.Float64Builder:
		return func(v reflect.Value) {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					e.AppendNull()
					return
				}
				v = v.Elem()
			}
			e.Append(v.Float())
		}
	case *array.BinaryDictionaryBuilder:
		return func(v reflect.Value) {
			if v.Kind() == reflect.Slice {
				if v.IsNil() {
					e.AppendNull()
					return
				}
				err := e.Append(v.Bytes())
				if err != nil {
					panic(err)
				}
				return
			}
			s := v.String()
			if s == "" {
				e.AppendNull()
			} else {
				e.Append([]byte(s))
			}
		}
	default:
		panic(fmt.Sprintf("%T is not supported builder", e))
	}
}

type Merger struct {
	b     *array.RecordBuilder
	merge func(arrow.Record)
}

func (m *Merger) Merge(record arrow.Record) {
	m.merge(record)
}

func (m *Merger) Add(record arrow.Record) {
	m.merge(record)
}

func (m *Merger) NewRecord(meta arrow.Metadata) arrow.Record {
	r := m.b.NewRecord()
	defer r.Release()
	return array.NewRecord(arrow.NewSchema(
		r.Schema().Fields(), &meta,
	), r.Columns(), r.NumRows())
}

func (m *Merger) Release() {
	m.b.Release()
}

func NewMerger(mem memory.Allocator, as *arrow.Schema) *Merger {
	b := array.NewRecordBuilder(mem, as)
	fields := make([]func(arrow.Array), len(b.Fields()))
	for i := range fields {
		fields[i] = merge(b.Field(i))
	}
	return &Merger{
		b: b,
		merge: func(r arrow.Record) {
			for i := 0; i < int(r.NumCols()); i++ {
				fields[i](r.Column(i))
			}
		},
	}
}

func merge(b array.Builder) func(arrow.Array) {
	switch e := b.(type) {
	case *array.BooleanBuilder:
		return func(v arrow.Array) {
			a := v.(*array.Boolean)
			e.Reserve(a.Len())
			for i := 0; i < a.Len(); i++ {
				if a.IsNull(i) {
					e.AppendNull()
					continue
				}
				e.UnsafeAppend(a.Value(i))
			}
		}
	case *array.Int64Builder:
		return func(v arrow.Array) {
			a := v.(*array.Int64)
			e.Reserve(a.Len())
			for i := 0; i < a.Len(); i++ {
				e.UnsafeAppend(a.Value(i))
			}
		}
	case *array.Float64Builder:
		return func(v arrow.Array) {
			a := v.(*array.Float64)
			e.Reserve(a.Len())
			for i := 0; i < a.Len(); i++ {
				e.UnsafeAppend(a.Value(i))
			}
		}
	case *array.BinaryDictionaryBuilder:
		return func(v arrow.Array) {
			a := v.(*array.Dictionary)
			x := a.Dictionary().(*array.String)
			for i := 0; i < a.Len(); i++ {
				if a.IsNull(i) {
					e.AppendNull()
					continue
				}
				e.AppendString(x.Value(a.GetValueIndex(i)))
			}
		}
	default:
		panic(fmt.Sprintf("%T is not supported builder", e))
	}
}

func NewTaker(mem memory.Allocator, projected []string) (*array.RecordBuilder, func(arrow.Record, []uint32)) {
	cols := make([]int, len(projected))
	fields := make([]arrow.Field, len(projected))
	for i, v := range projected {
		cols[i] = events.Mapping[v]
		fields[i] = events.Schema.Field(events.Mapping[v])
	}
	b := array.NewRecordBuilder(mem, arrow.NewSchema(fields, nil))
	tf := make([]func(arrow.Array, []uint32), len(projected))
	for i, f := range b.Fields() {
		tf[i] = take(f)
	}
	return b, func(v arrow.Record, rows []uint32) {
		for idx, col := range cols {
			tf[idx](v.Column(col), rows)
		}
	}
}

func take(b array.Builder) func(arrow.Array, []uint32) {
	switch e := b.(type) {
	case *array.BooleanBuilder:
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.Boolean)
			e.Reserve(len(rows))
			for _, i := range rows {
				if a.IsNull(int(i)) {
					e.AppendNull()
					continue
				}
				e.UnsafeAppend(a.Value(int(i)))
			}
		}
	case *array.Int64Builder:
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.Int64)
			e.Reserve(len(rows))
			for _, i := range rows {
				e.UnsafeAppend(a.Value(int(i)))
			}
		}
	case *array.Float64Builder:
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.Float64)
			e.Reserve(len(rows))
			for _, i := range rows {
				e.UnsafeAppend(a.Value(int(i)))
			}
		}
	case *array.BinaryDictionaryBuilder:
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.Dictionary)
			x := a.Dictionary().(*array.String)
			for _, i := range rows {
				if a.IsNull(int(i)) {
					e.AppendNull()
					continue
				}
				e.AppendString(x.Value(a.GetValueIndex(int(i))))
			}
		}
	default:
		panic(fmt.Sprintf("%T is not supported builder", e))
	}
}
