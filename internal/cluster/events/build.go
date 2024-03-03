package events

import (
	"fmt"
	"sort"
	"sync"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/memory"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
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

type Multi struct {
	mu     sync.Mutex
	mem    memory.Allocator
	builds map[string]*Builder
}

func NewMulti(mem memory.Allocator) *Multi {
	return &Multi{mem: mem, builds: make(map[string]*Builder)}
}

func (m *Multi) Append(e *v1.Data) {
	m.mu.Lock()
	b, ok := m.builds[e.TenantId]
	if !ok {
		b = New(m.mem, map[string]string{
			"tenant_id": e.TenantId,
		})
		m.builds[e.TenantId] = b
	}
	b.WriteData(e)
	PutOne(e)
	m.mu.Unlock()
}

func (m *Multi) All(f func(tenantId string, r arrow.Record)) {
	m.mu.Lock()
	for k, v := range m.builds {
		f(k, v.NewRecord())
	}
	m.mu.Unlock()
}

type Builder struct {
	r      *array.RecordBuilder
	fields map[protoreflect.FieldNumber]buildFunc
}

type field struct {
	number protoreflect.FieldNumber
	arrow  arrow.Field
}

func New(mem memory.Allocator, metadata ...map[string]string) *Builder {
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
	var meta *arrow.Metadata
	if len(metadata) > 0 {
		m := arrow.MetadataFrom(metadata[0])
		meta = &m
	}
	r := array.NewRecordBuilder(mem, arrow.NewSchema(af, meta))
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
func (b *Builder) Len() int {
	return b.r.Field(0).Len()
}

func (b *Builder) Write(list *v1.Data_List) arrow.Record {
	ls := list.GetItems()
	sort.SliceStable(ls, func(i, j int) bool {
		return ls[i].Timestamp < ls[j].Timestamp
	})
	b.r.Reserve(len(ls))
	for _, e := range ls {
		b.WriteData(e)
	}
	return b.NewRecord()
}

func (b *Builder) WriteData(data *v1.Data) {
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
