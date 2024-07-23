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

func (s *Schema[T]) Release() {
	s.msg.builder.Release()
}

type valueFn func(protoreflect.Value) error

type node struct {
	field arrow.Field
	setup func(array.Builder) valueFn
	write valueFn
}

func build(msg protoreflect.Message) *message {
	fields := msg.Descriptor().Fields()
	nodes := make([]*node, fields.Len())
	a := make([]arrow.Field, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		nodes[i] = newNode(fields.Get(i))
		a[i] = nodes[i].field
	}
	as := arrow.NewSchema(a, nil)
	return &message{
		fields: nodes,
		schema: as,
	}
}

type message struct {
	fields  []*node
	schema  *arrow.Schema
	builder *array.RecordBuilder
}

func (m *message) build(mem memory.Allocator) {
	b := array.NewRecordBuilder(mem, m.schema)
	for i, ch := range m.fields {
		ch.build(b.Field(i))
	}
	m.builder = b
}

func (m *message) append(msg protoreflect.Message) {
	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		m.fields[i].write(msg.Get(fields.Get(i)))
	}
}

func (m *message) NewRecord() arrow.Record {
	return m.builder.NewRecord()
}

func (n *node) build(a array.Builder) {
	n.write = n.setup(a)
}

func newNode(field protoreflect.FieldDescriptor) *node {
	switch field.Kind() {
	case protoreflect.BoolKind:
		return &node{
			field: arrow.Field{
				Name: string(field.Name()),
				Type: arrow.FixedWidthTypes.Boolean,
			},
			setup: func(b array.Builder) valueFn {
				a := b.(*array.BooleanBuilder)
				return func(v protoreflect.Value) error {
					a.Append(v.Bool())
					return nil
				}
			},
		}
	case protoreflect.Uint64Kind:
		return &node{
			field: arrow.Field{
				Name: string(field.Name()),
				Type: arrow.PrimitiveTypes.Uint64,
			},
			setup: func(b array.Builder) valueFn {
				a := b.(*array.Uint64Builder)
				return func(v protoreflect.Value) error {
					a.Append(v.Uint())
					return nil
				}
			},
		}
	default:
		panic(fmt.Sprintf("%v is not supported", field.Kind()))
	}
}
