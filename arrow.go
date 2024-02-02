package staples

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/memory"
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
				Name:     f.Name,
				Type:     base,
				Nullable: f.Type.Kind() == reflect.Ptr || typ.Kind() == reflect.String,
			})
			continue
		}
		if typ == byteArray {
			o = append(o, arrow.Field{
				Name:     f.Name,
				Type:     baseTypes[reflect.String],
				Nullable: true,
			})
			continue
		}
		switch typ.Kind() {
		case reflect.Slice:
			typ = typ.Elem()
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			if base, ok := baseTypes[typ.Kind()]; ok {
				o = append(o, arrow.Field{
					Name:     f.Name,
					Type:     arrow.ListOf(base),
					Nullable: true,
				})
				continue
			}
			if typ.Kind() == reflect.Struct {
				o = append(o, arrow.Field{
					Name:     f.Name,
					Type:     arrow.ListOf(arrow.StructOf(build(typ)...)),
					Nullable: true,
				})
				continue
			}
			panic(typ.String() + " slices are not supported")
		case reflect.Struct:
			o = append(o, arrow.Field{
				Name:     f.Name,
				Type:     arrow.StructOf(build(typ)...),
				Nullable: f.Type.Kind() == reflect.Ptr,
			})
		case reflect.Map:
			o = append(o, arrow.Field{
				Name: f.Name,
				Type: arrow.MapOf(
					baseTypes[reflect.String],
					baseTypes[reflect.String],
				),
				Nullable: true,
			})
		default:
			panic(typ.String() + " is not supported")
		}
	}
	return
}

var baseTypes = map[reflect.Kind]arrow.DataType{
	reflect.Bool: arrow.FixedWidthTypes.Boolean,
	reflect.String: &arrow.DictionaryType{
		IndexType: arrow.PrimitiveTypes.Uint32,
		ValueType: arrow.BinaryTypes.String,
	},
	reflect.Int32:   arrow.PrimitiveTypes.Int32,
	reflect.Uint32:  arrow.PrimitiveTypes.Uint32,
	reflect.Int64:   arrow.PrimitiveTypes.Int64,
	reflect.Uint64:  arrow.PrimitiveTypes.Uint64,
	reflect.Float64: arrow.PrimitiveTypes.Float64,
}

var byteArray = reflect.TypeOf([]byte{})

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
	case *array.Int32Builder:
		return func(v reflect.Value) {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					e.AppendNull()
					return
				}
				v = v.Elem()
			}
			e.Append(int32(v.Int()))
		}
	case *array.Uint32Builder:
		return func(v reflect.Value) {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					e.AppendNull()
					return
				}
				v = v.Elem()
			}
			e.Append(uint32(v.Uint()))
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
	case *array.Uint64Builder:
		return func(v reflect.Value) {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					e.AppendNull()
					return
				}
				v = v.Elem()
			}
			e.Append(uint64(v.Uint()))
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

	case *array.MapBuilder:
		key := e.KeyBuilder().(*array.BinaryDictionaryBuilder)
		value := e.ItemBuilder().(*array.BinaryDictionaryBuilder)
		return func(v reflect.Value) {
			if v.IsNil() {
				e.AppendNull()
				return
			}
			e.Append(true)
			for _, k := range v.MapKeys() {
				key.Append([]byte(k.String()))
				value.Append([]byte(v.MapIndex(k).String()))
			}
		}
	case *array.StructBuilder:
		fields := make([]func(reflect.Value), e.NumField())
		for i := range fields {
			fields[i] = write(e.FieldBuilder(i))
		}
		return func(v reflect.Value) {
			if v.Kind() == reflect.Ptr {
				if v.IsNil() {
					e.AppendNull()
					return
				}
				v = v.Elem()
			}
			e.Append(true)
			for i := 0; i < v.NumField(); i++ {
				fields[i](v.Field(i))
			}
		}
	case *array.ListBuilder:
		value := write(e.ValueBuilder())
		return func(v reflect.Value) {
			if v.IsNil() || v.Len() == 0 {
				e.AppendNull()
				return
			}
			e.Append(true)
			for i := 0; i < v.Len(); i++ {
				value(v.Index(i))
			}
		}
	default:
		panic(fmt.Sprintf("%T is not supported builder", e))
	}
}

type Merger struct {
	mu    sync.Mutex
	b     *array.RecordBuilder
	merge func(arrow.Record)
}

func (m *Merger) Merge(a arrow.Record) {
	m.mu.Lock()
	m.merge(a)
	m.mu.Unlock()
}

func (m *Merger) NewRecord() arrow.Record {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.b.NewRecord()
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
	case *array.Int32Builder:
		return func(v arrow.Array) {
			a := v.(*array.Int32)
			e.Reserve(a.Len())
			for i := 0; i < a.Len(); i++ {
				e.UnsafeAppend(a.Value(i))
			}
		}
	case *array.Uint32Builder:
		return func(v arrow.Array) {
			a := v.(*array.Uint32)
			e.Reserve(a.Len())
			for i := 0; i < a.Len(); i++ {
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
	case *array.Uint64Builder:
		return func(v arrow.Array) {
			a := v.(*array.Uint64)
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
	case *array.BooleanBuilder:
		return func(v arrow.Array) {
			a := v.(*array.Boolean)
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
	case *array.StructBuilder:
		fields := make([]func(arrow.Array), e.NumField())
		for i := range fields {
			fields[i] = merge(e.FieldBuilder(i))
		}
		return func(v arrow.Array) {
			a := v.(*array.Struct)
			e.Reserve(a.Len())
			for i := 0; i < a.Len(); i++ {
				if a.IsNull(i) {
					e.UnsafeAppendBoolToBitmap(false)
				} else {
					e.UnsafeAppendBoolToBitmap(true)
				}
			}
			for i := range fields {
				fields[i](a.Field(i))
			}
		}
	case *array.ListBuilder:
		value := merge(e.ValueBuilder())
		return func(v arrow.Array) {
			a := v.(*array.List)
			x := a.ListValues()
			for i := 0; i < a.Len(); i++ {
				if a.IsNull(i) {
					e.AppendNull()
					continue
				}
				start, end := a.ValueOffsets(i)
				chunk := array.NewSlice(x, start, end)
				e.Append(true)
				value(chunk)
				chunk.Release()
			}
		}
	default:
		panic(fmt.Sprintf("%T is not supported builder", e))
	}
}

func NewTaker(mem memory.Allocator, as *arrow.Schema) (*array.RecordBuilder, func(arrow.Record, []int, []uint32)) {
	b := array.NewRecordBuilder(mem, as)
	fields := make([]func(arrow.Array, []uint32), len(b.Fields()))
	for i := range fields {
		fields[i] = take(b.Field(i))
	}
	return b, func(v arrow.Record, columns []int, rows []uint32) {
		for _, i := range columns {
			fields[i](v.Column(i), rows)
		}
	}
}
func take(b array.Builder) func(arrow.Array, []uint32) {
	switch e := b.(type) {
	case *array.Int32Builder:
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.Int32)
			e.Reserve(len(rows))
			for _, i := range rows {
				e.UnsafeAppend(a.Value(int(i)))
			}
		}
	case *array.Uint32Builder:
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.Uint32)
			e.Reserve(len(rows))
			for _, i := range rows {
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
	case *array.Uint64Builder:
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.Uint64)
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
	case *array.BooleanBuilder:
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.Boolean)
			e.Reserve(len(rows))
			for _, i := range rows {
				e.UnsafeAppend(a.Value(int(i)))
			}
		}
	case *array.BinaryDictionaryBuilder:
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.Dictionary)
			x := a.Dictionary().(*array.Binary)
			for _, i := range rows {
				if a.IsNull(int(i)) {
					e.AppendNull()
					continue
				}
				e.Append(x.Value(a.GetValueIndex(int(i))))
			}
		}
	case *array.StructBuilder:
		fields := make([]func(arrow.Array, []uint32), e.NumField())
		for i := range fields {
			fields[i] = take(e.FieldBuilder(i))
		}
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.Struct)
			e.Reserve(len(rows))
			for _, i := range rows {
				if a.IsNull(int(i)) {
					e.UnsafeAppendBoolToBitmap(false)
				} else {
					e.UnsafeAppendBoolToBitmap(true)
				}
			}
			for i := range fields {
				fields[i](a.Field(i), rows)
			}
		}
	case *array.ListBuilder:
		value := merge(e.ValueBuilder())
		return func(v arrow.Array, rows []uint32) {
			a := v.(*array.List)
			x := a.ListValues()
			for _, i := range rows {
				if a.IsNull(int(i)) {
					e.AppendNull()
					continue
				}
				start, end := a.ValueOffsets(int(i))
				chunk := array.NewSlice(x, start, end)
				e.Append(true)
				value(chunk)
				chunk.Release()
			}
		}
	default:
		panic(fmt.Sprintf("%T is not supported builder", e))
	}
}

func IterAttributes(a *array.Struct, f func(key string, value string)) {
	kd := a.Field(0).(*array.Dictionary)
	keys := kd.Dictionary().(*array.String)
	vd := a.Field(1).(*array.Dictionary)
	values := vd.Dictionary().(*array.String)
	for i := 0; i < a.Len(); i++ {
		if a.IsNull(i) {
			continue
		}
		f(keys.Value(kd.GetValueIndex(i)), values.Value(vd.GetValueIndex(i)))
	}
}
