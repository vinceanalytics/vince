package len64

import (
	"errors"
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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
		x := createNode(root, fields.Get(i))
		root.children[i] = x
		root.hash[x.field.Name] = x
		a[i] = root.children[i].field
	}
	as := arrow.NewSchema(a, nil)
	return &message{
		root:   root,
		schema: as,
	}
}

type message struct {
	root    *node
	schema  *arrow.Schema
	builder *array.RecordBuilder
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
func createNode(parent *node, field protoreflect.FieldDescriptor) *node {
	n := &node{parent: parent, desc: field, field: arrow.Field{
		Name:     string(field.Name()),
		Nullable: nullable(field),
	}, hash: make(map[string]*node)}
	n.field.Type = n.baseType(field)

	if n.field.Type != nil {
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
	case protoreflect.Uint64Kind:
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
