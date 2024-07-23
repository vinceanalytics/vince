package len64

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/apache/arrow/go/v18/parquet"
	"github.com/apache/arrow/go/v18/parquet/compress"
	"github.com/apache/arrow/go/v18/parquet/pqarrow"
	"github.com/apache/arrow/go/v18/parquet/schema"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	maxDepth = 10
)

var (
	otelAnyDescriptor = (&commonv1.AnyValue{}).ProtoReflect().Descriptor()
)

var ErrMxDepth = errors.New("max depth reached, either the message is deeply nested or a circular dependency was introduced")

type Schema[T proto.Message] struct {
	msg *message
}

func newSchema[T proto.Message](mem memory.Allocator) (schema *Schema[T], err error) {
	defer func() {
		e := recover()
		if e != nil {
			switch x := e.(type) {
			case error:
				err = x
			case string:
				err = errors.New(x)
			default:
				panic(x)
			}
		}
	}()
	var a T
	b := build(a.ProtoReflect())
	b.build(mem)
	schema = &Schema[T]{msg: b}
	return
}

// Append appends protobuf value to the schema builder.This method is not safe
// for concurrent use.
func (s *Schema[T]) Append(value T) {
	s.msg.append(value.ProtoReflect())
}

// NewRecord returns buffered builder value as an arrow.Record. The builder is
// reset and can be reused to build new records.
func (s *Schema[T]) NewRecord() arrow.Record {
	return s.msg.NewRecord()
}

// Parquet returns schema as arrow schema
func (s *Schema[T]) Schema() *arrow.Schema {
	return s.msg.schema
}

// One decodes value at row from r into a field with name. Useful for reading
// nested list/struct messages.
//
// FOr scalar values it is much faster  to read directly from the array.
func (s *Schema[T]) One(msg T, name string, r arrow.Array, row int) {
	proto.Reset(msg)
	if r.IsNull(row) {
		return
	}
	nx, ok := s.msg.root.hash[name]
	if !ok {
		return
	}
	fs := nx.desc.(protoreflect.FieldDescriptor)

	switch {
	case fs.IsList():

		ls := r.(*array.List)
		start, end := ls.ValueOffsets(row)
		val := ls.ListValues()
		lv := msg.ProtoReflect().NewField(fs)
		list := lv.List()
		for k := start; k < end; k++ {
			list.Append(
				nx.encode(
					list.NewElement(),
					val,
					int(k),
				),
			)
		}
		msg.ProtoReflect().Set(fs, lv)
	case fs.IsMap():
		panic("MAP not supported")
	default:
		lv := msg.ProtoReflect().NewField(fs)
		nx.encode(lv, r, row)
		msg.ProtoReflect().Set(fs, lv)
	}
}

func (s *Schema[T]) Release() {
	s.msg.builder.Release()
}

type valueFn func(protoreflect.Value, bool) error

type encodeFn func(value protoreflect.Value, a arrow.Array, row int) protoreflect.Value

type node struct {
	parent   *node
	field    arrow.Field
	setup    func(array.Builder) valueFn
	write    valueFn
	desc     protoreflect.Descriptor
	children []*node
	encode   encodeFn
	hash     map[string]*node
}

func build(msg protoreflect.Message) *message {
	root := &node{desc: msg.Descriptor(),
		field: arrow.Field{},
		hash:  make(map[string]*node),
	}
	fields := msg.Descriptor().Fields()
	root.children = make([]*node, fields.Len())
	a := make([]arrow.Field, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		x := createNode(root, fields.Get(i), 0)
		root.children[i] = x
		root.hash[x.field.Name] = x
		a[i] = root.children[i].field
	}
	as := arrow.NewSchema(a, nil)

	// we need to apply compression on all fields and use dictionary for binary and
	// string columns.
	bs, err := pqarrow.ToParquet(as, parquet.NewWriterProperties(), pqarrow.DefaultWriterProps())
	if err != nil {
		panic(err)
	}
	var props []parquet.WriterProperty

	for i := 0; i < bs.NumColumns(); i++ {
		col := bs.Column(i)
		if col.PhysicalType() == parquet.Types.ByteArray {
			props = append(props, parquet.WithDictionaryPath(col.ColumnPath(), true))
		}
	}
	// ZSTD is pretty good for all cases. Default level is reasonable.
	props = append(props, parquet.WithCompression(compress.Codecs.Zstd))
	// All writes are on a single row group. This is needed because we treat rows
	// are sample ID and we need to keep the mapping
	props = append(props, parquet.WithMaxRowGroupLength(math.MaxInt))

	ps, err := pqarrow.ToParquet(as, parquet.NewWriterProperties(props...), pqarrow.DefaultWriterProps())
	if err != nil {
		panic(err)
	}
	return &message{
		root:    root,
		schema:  as,
		parquet: ps,
		props:   props,
	}
}

type message struct {
	root    *node
	schema  *arrow.Schema
	parquet *schema.Schema
	builder *array.RecordBuilder
	props   []parquet.WriterProperty
}

func (m *message) build(mem memory.Allocator) {
	b := array.NewRecordBuilder(mem, m.schema)
	for i, ch := range m.root.children {
		ch.build(b.Field(i))
	}
	m.builder = b
}

func (m *message) append(msg protoreflect.Message) {
	m.root.WriteMessage(msg)
}

func (m *message) NewRecord() arrow.Record {
	return m.builder.NewRecord()
}
func createNode(parent *node, field protoreflect.FieldDescriptor, depth int) *node {
	if depth >= maxDepth {
		panic(ErrMxDepth)
	}
	name, ok := parent.field.Metadata.GetValue("path")
	if ok {
		name += "." + string(field.Name())
	} else {
		name = string(field.Name())
	}
	n := &node{parent: parent, desc: field, field: arrow.Field{
		Name:     string(field.Name()),
		Nullable: nullable(field),
		Metadata: arrow.MetadataFrom(map[string]string{
			"path":             name,
			"PARQUET:field_id": strconv.Itoa(int(field.Number())),
		}),
	}, hash: make(map[string]*node)}
	n.field.Type = n.baseType(field)

	if n.field.Type != nil {
		return n
	}
	// Try a message
	if msg := field.Message(); msg != nil {
		switch msg {
		case otelAnyDescriptor:
			n.field.Type = arrow.BinaryTypes.Binary
			n.field.Nullable = true
			n.setup = func(b array.Builder) valueFn {
				a := b.(*array.BinaryBuilder)
				return func(v protoreflect.Value, set bool) error {
					if !v.IsValid() {
						a.AppendNull()
						return nil
					}
					e := v.Message().Interface().(*commonv1.AnyValue)
					bs, err := proto.Marshal(e)
					if err != nil {
						return err
					}
					a.Append(bs)
					return nil
				}
			}
			n.encode = func(value protoreflect.Value, a arrow.Array, row int) protoreflect.Value {
				if a.IsNull(row) {
					return protoreflect.Value{}
				}
				msg := value.Message()
				var v []byte
				if a.DataType().ID() == arrow.DICTIONARY {
					d := a.(*array.Dictionary)
					v = d.Dictionary().(*array.Binary).Value(d.GetValueIndex(row))
				} else {
					v = a.(*array.Binary).Value(row)
				}
				proto.Unmarshal(v, msg.Interface())
				return value
			}
		}
		if n.field.Type != nil {
			if field.IsList() {
				n.field.Type = arrow.ListOf(n.field.Type)
				setup := n.setup
				n.setup = func(b array.Builder) valueFn {
					ls := b.(*array.ListBuilder)
					value := setup(ls.ValueBuilder())
					return func(v protoreflect.Value, set bool) error {
						if !v.IsValid() {
							ls.AppendNull()
							return nil
						}
						ls.Append(true)
						list := v.List()
						for i := 0; i < list.Len(); i++ {
							err := value(list.Get(i), true)
							if err != nil {
								return err
							}
						}
						return nil
					}
				}
			}
			return n
		}
		f := msg.Fields()
		n.children = make([]*node, f.Len())
		a := make([]arrow.Field, f.Len())
		for i := 0; i < f.Len(); i++ {
			x := createNode(n, f.Get(i), depth+1)
			n.children[i] = x
			n.hash[x.field.Name] = x
			a[i] = n.children[i].field
		}
		n.field.Type = arrow.StructOf(a...)
		n.field.Nullable = true
		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.StructBuilder)
			fs := make([]valueFn, len(n.children))
			for i := range n.children {
				fs[i] = n.children[i].setup(a.FieldBuilder(i))
			}
			return func(v protoreflect.Value, set bool) error {
				if !v.IsValid() {
					a.AppendNull()
					return nil
				}
				a.Append(true)
				msg := v.Message()
				fields := msg.Descriptor().Fields()
				for i := 0; i < fields.Len(); i++ {
					err := fs[i](msg.Get(fields.Get(i)), msg.Has(fields.Get(i)))
					if err != nil {
						return err
					}
				}
				return nil
			}
		}
		n.encode = func(value protoreflect.Value, a arrow.Array, row int) protoreflect.Value {
			msg := value.Message()
			s := a.(*array.Struct)
			typ := a.DataType().(*arrow.StructType)
			for j := 0; j < s.NumField(); j++ {
				f := typ.Field(j)
				nx, ok := n.hash[f.Name]
				if !ok {
					panic(fmt.Sprintf("arrow3: field %s not found in node %v", f.Name, n.field.Name))
				}
				if s.Field(j).IsNull(row) {
					continue
				}
				fs := nx.desc.(protoreflect.FieldDescriptor)
				switch {
				case fs.IsList():
					ls := s.Field(j).(*array.List)
					start, end := ls.ValueOffsets(row)
					if start != end {
						lv := msg.Mutable(fs)
						list := lv.List()
						va := ls.ListValues()
						for k := start; k < end; k++ {
							list.Append(
								nx.encode(list.NewElement(), va, int(k)),
							)
						}
						msg.Set(fs, lv)
					}
				case fs.IsMap():
					panic("MAP not supported")
				default:
					msg.Set(fs, nx.encode(msg.NewField(fs), s.Field(j), row))
				}
			}
			return value
		}
		if field.IsList() {
			n.field.Type = arrow.ListOf(n.field.Type)
			setup := n.setup
			n.setup = func(b array.Builder) valueFn {
				ls := b.(*array.ListBuilder)
				value := setup(ls.ValueBuilder())
				return func(v protoreflect.Value, set bool) error {
					if !v.IsValid() {
						ls.AppendNull()
						return nil
					}
					ls.Append(true)
					list := v.List()
					for i := 0; i < list.Len(); i++ {
						err := value(list.Get(i), true)
						if err != nil {
							return err
						}
					}
					return nil
				}

			}
		}
		if field.ContainingOneof() != nil {
			setup := n.setup
			n.setup = func(b array.Builder) valueFn {
				do := setup(b)
				return func(v protoreflect.Value, set bool) error {
					if !set {
						b.AppendNull()
						return nil
					}
					return do(v, set)
				}
			}
		}
		return n
	}
	panic(fmt.Sprintf("%v is not supported ", field.Name()))
}

func (n *node) build(a array.Builder) {
	n.write = n.setup(a)
}

func (n *node) WriteMessage(msg protoreflect.Message) {
	f := msg.Descriptor().Fields()
	for i := 0; i < f.Len(); i++ {
		n.children[i].write(msg.Get(f.Get(i)), msg.Has(f.Get(i)))
	}
}

func (n *node) baseType(field protoreflect.FieldDescriptor) (t arrow.DataType) {
	switch field.Kind() {
	case protoreflect.EnumKind:
		t = arrow.PrimitiveTypes.Int32

		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.Int32Builder)
			return func(v protoreflect.Value, set bool) error {
				a.Append(int32(v.Enum()))
				return nil
			}
		}
		n.encode = func(value protoreflect.Value, a arrow.Array, i int) protoreflect.Value {
			return protoreflect.ValueOfEnum(
				protoreflect.EnumNumber(
					a.(*array.Int32).Value(i),
				),
			)
		}
	case protoreflect.BoolKind:
		t = arrow.FixedWidthTypes.Boolean

		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.BooleanBuilder)
			return func(v protoreflect.Value, set bool) error {
				a.Append(v.Bool())
				return nil
			}
		}
		n.encode = func(value protoreflect.Value, a arrow.Array, i int) protoreflect.Value {
			return protoreflect.ValueOfBool(a.(*array.Boolean).Value(i))
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		t = arrow.PrimitiveTypes.Int32
		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.Int32Builder)
			return func(v protoreflect.Value, set bool) error {
				a.Append(int32(v.Int()))
				return nil
			}
		}
		n.encode = func(value protoreflect.Value, a arrow.Array, i int) protoreflect.Value {
			return protoreflect.ValueOfInt32(a.(*array.Int32).Value(i))
		}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.Uint32Builder)
			return func(v protoreflect.Value, set bool) error {
				a.Append(uint32(v.Uint()))
				return nil
			}
		}
		t = arrow.PrimitiveTypes.Uint32
		n.encode = func(value protoreflect.Value, a arrow.Array, i int) protoreflect.Value {
			return protoreflect.ValueOfUint32(a.(*array.Uint32).Value(i))
		}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.Int64Builder)
			return func(v protoreflect.Value, set bool) error {
				a.Append(v.Int())
				return nil
			}
		}
		t = arrow.PrimitiveTypes.Int64
		n.encode = func(value protoreflect.Value, a arrow.Array, i int) protoreflect.Value {
			return protoreflect.ValueOfInt64(a.(*array.Int64).Value(i))
		}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.Uint64Builder)
			return func(v protoreflect.Value, set bool) error {
				a.Append(v.Uint())
				return nil
			}
		}
		t = arrow.PrimitiveTypes.Uint64
		n.encode = func(value protoreflect.Value, a arrow.Array, i int) protoreflect.Value {
			return protoreflect.ValueOfUint64(a.(*array.Uint64).Value(i))
		}
	case protoreflect.DoubleKind:
		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.Float64Builder)
			return func(v protoreflect.Value, set bool) error {
				a.Append(v.Float())
				return nil
			}
		}
		t = arrow.PrimitiveTypes.Float64
		n.encode = func(value protoreflect.Value, a arrow.Array, i int) protoreflect.Value {
			return protoreflect.ValueOfFloat64(a.(*array.Float64).Value(i))
		}
	case protoreflect.FloatKind:
		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.Float32Builder)
			return func(v protoreflect.Value, set bool) error {
				a.Append(float32(v.Float()))
				return nil
			}
		}
		t = arrow.PrimitiveTypes.Float32
		n.encode = func(value protoreflect.Value, a arrow.Array, i int) protoreflect.Value {
			return protoreflect.ValueOfFloat32(a.(*array.Float32).Value(i))
		}
	case protoreflect.StringKind:
		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.StringBuilder)
			return func(v protoreflect.Value, set bool) error {
				a.Append(v.String())
				return nil
			}
		}
		t = arrow.BinaryTypes.String
		n.encode = func(value protoreflect.Value, a arrow.Array, i int) protoreflect.Value {
			if a.DataType().ID() == arrow.DICTIONARY {
				d := a.(*array.Dictionary)
				return protoreflect.ValueOfString(
					d.Dictionary().(*array.String).Value(d.GetValueIndex(i)),
				)
			}
			return protoreflect.ValueOfString(a.(*array.String).Value(i))
		}
	case protoreflect.BytesKind:
		n.setup = func(b array.Builder) valueFn {
			a := b.(*array.BinaryBuilder)
			return func(v protoreflect.Value, set bool) error {
				if !v.IsValid() {
					a.AppendNull()
					return nil
				}
				a.Append(v.Bytes())
				return nil
			}
		}
		t = arrow.BinaryTypes.Binary
		n.encode = func(value protoreflect.Value, a arrow.Array, i int) protoreflect.Value {
			if a.DataType().ID() == arrow.DICTIONARY {
				d := a.(*array.Dictionary)
				return protoreflect.ValueOfBytes(
					d.Dictionary().(*array.Binary).Value(d.GetValueIndex(i)),
				)
			}
			return protoreflect.ValueOfBytes(a.(*array.Binary).Value(i))
		}
	}
	if field.IsList() {
		if t != nil {
			setup := n.setup
			n.setup = func(b array.Builder) valueFn {
				ls := b.(*array.ListBuilder)
				vb := setup(ls.ValueBuilder())
				return func(v protoreflect.Value, set bool) error {
					if !v.IsValid() {
						ls.AppendNull()
						return nil
					}
					ls.Append(true)
					list := v.List()
					for i := 0; i < list.Len(); i++ {
						err := vb(list.Get(i), true)
						if err != nil {
							return err
						}
					}
					return nil
				}
			}
			t = arrow.ListOf(t)
		}
	}
	if t != nil && field.ContainingOneof() != nil {
		// Handle oneof for base types
		setup := n.setup
		n.setup = func(b array.Builder) valueFn {
			do := setup(b)
			return func(v protoreflect.Value, set bool) error {
				if !set {
					b.AppendNull()
					return nil
				}
				return do(v, set)
			}
		}

	}
	if field.IsMap() {
		panic("MAP not supported")
	}
	return
}

func nullable(f protoreflect.FieldDescriptor) bool {
	return f.HasOptionalKeyword() || f.ContainingOneof() != nil ||
		f.Kind() == protoreflect.BytesKind
}
