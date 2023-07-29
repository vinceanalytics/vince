package neo

import (
	"context"
	"sort"
	"sync"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/compute"
	"github.com/apache/arrow/go/v13/arrow/scalar"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/blocks"
	"github.com/vinceanalytics/vince/pkg/entry"
)

type Analysis interface {
	ColumnIndices() []int
	Analyze(context.Context, arrow.Record)
}

var _ Analysis = (*Base)(nil)

type Base struct {
	columns []string
	indices []int
	filters []*blocks.Filter
	build   *array.RecordBuilder
}

func NewBase(pick []string, filters ...*blocks.Filter) *Base {
	b := basePool.Get().(*Base)
	return b.Init(pick, filters...)
}

var basePool = &sync.Pool{
	New: func() any {
		return &Base{
			columns: make([]string, 0, len(entry.Index)),
			indices: make([]int, 0, len(entry.Index)),
			filters: make([]*blocks.Filter, 0, len(entry.Index)),
		}
	},
}

func (b *Base) ColumnIndices() []int {
	return b.indices
}

func (b *Base) Select() []string {
	return b.columns
}

func (b *Base) Record() arrow.Record {
	return b.build.NewRecord()
}

func (b *Base) Init(pick []string, filters ...*blocks.Filter) *Base {
	b.columns = append(b.columns,
		"bounce",
		"duration",
		"id",
		"name",
		"timestamp",
	)

	seen := map[string]struct{}{
		"bounce":    {},
		"duration":  {},
		"id":        {},
		"name":      {},
		"timestamp": {},
	}
	for i := range pick {
		_, ok := seen[pick[i]]
		if ok {
			continue
		}
		b.columns = append(b.columns, pick[i])
		seen[pick[i]] = struct{}{}
	}
	for i := range filters {
		_, ok := seen[filters[i].Column]
		if ok {
			continue
		}
		b.columns = append(b.columns, filters[i].Column)
		seen[filters[i].Column] = struct{}{}
	}
	sort.Strings(b.columns)

	for i := range b.columns {
		b.indices = append(b.indices, entry.Index[b.columns[i]])
	}
	b.filters = append(b.filters, filters...)
	return b
}

func (b *Base) Release() {
	b.columns = b.columns[:0]
	b.indices = b.indices[:0]
	b.filters = b.filters[:0]
	if b.build != nil {
		b.build.Release()
		b.build = nil
	}
	basePool.Put(b)
}

func (b *Base) Analyze(ctx context.Context, r arrow.Record) {
	b.selection(ctx, r)
}

func (b *Base) selection(ctx context.Context, r arrow.Record) {
	defer r.Release()
	columns := make(map[string]arrow.Array)
	for i := 0; i < int(r.NumCols()); i++ {
		columns[r.ColumnName(i)] = r.Column(i)
	}
	var activeFilter arrow.Array
	for _, f := range b.filters {
		o := apply(ctx, f, columns[f.Column])
		if activeFilter != nil {
			n := call(ctx, "and",
				compute.NewDatum(activeFilter),
				compute.NewDatum(o),
			)
			o.Release()
			activeFilter.Release()
			activeFilter = n
		} else {
			activeFilter = o
		}
	}
	schema := b.schema(r)
	cols := make([]arrow.Array, len(columns))
	for i := range b.columns {
		cols[i] = columns[b.columns[i]]
	}
	record := array.NewRecord(schema, cols, r.NumRows())
	r.Release()
	if activeFilter != nil {
		o := filter(ctx,
			compute.NewDatum(record),
			compute.NewDatum(activeFilter),
		)
		activeFilter.Release()
		b.merge(o)
		return
	}
	b.merge(record)
}

func (b *Base) merge(r arrow.Record) {
	defer r.Release()
	f := b.build.Fields()
	for i := range f {
		switch e := f[i].(type) {
		case *array.TimestampBuilder:
			a := r.Column(i).(*array.Timestamp)
			e.AppendValues(a.TimestampValues(), nil)
		case *array.StringBuilder:
			a := r.Column(i).(*array.String)
			e.Reserve(a.Len())
			for n := 0; n < a.Len(); n++ {
				e.Append(a.Value(n))
			}
		case *array.Int64Builder:
			a := r.Column(i).(*array.Int64)
			e.AppendValues(a.Int64Values(), nil)
		}
	}
}

func (b *Base) schema(r arrow.Record) *arrow.Schema {
	if b.build == nil {
		s := r.Schema()
		pick := b.Select()
		f := make([]arrow.Field, 0, len(pick))
		sort.Strings(pick)
		for _, n := range pick {
			f = append(f, arrow.Field{
				Name: n,
				Type: s.Field(s.FieldIndices(n)[0]).Type,
			})
		}
		b.build = array.NewRecordBuilder(entry.Pool, arrow.NewSchema(f, nil))
	}
	return b.build.Schema()
}

func apply(ctx context.Context, f *blocks.Filter, a arrow.Array) arrow.Array {
	switch e := f.Value.(type) {
	case *blocks.Filter_Str:
		return call(ctx, f.Op.String(),
			compute.NewDatum(a),
			compute.NewDatum(scalar.NewStringScalar(e.Str)),
		)
	case *blocks.Filter_Timestamp:
		return call(ctx, f.Op.String(),
			compute.NewDatum(a),
			compute.NewDatum(scalar.NewTimestampScalar(arrow.Timestamp(e.Timestamp),
				&arrow.TimestampType{Unit: arrow.Millisecond})),
		)
	case *blocks.Filter_Duration:
		return call(ctx, f.Op.String(),
			compute.NewDatum(a),
			compute.NewDatum(scalar.NewDurationScalar(arrow.Duration(e.Duration),
				&arrow.DurationType{Unit: arrow.Nanosecond})),
		)
	default:
		panic("unreachable")
	}
}

func call(ctx context.Context, f string, a ...compute.Datum) arrow.Array {
	o := must.Must(compute.CallFunction(ctx, f, nil, a...))
	r := o.(*compute.ArrayDatum).MakeArray()
	for i := range a {
		a[i].Release()
		a[i] = nil
	}
	o.Release()
	return r
}

func filter(ctx context.Context, a ...compute.Datum) arrow.Record {
	o := must.Must(compute.CallFunction(ctx, "filter", compute.DefaultFilterOptions(), a...))
	r := o.(*compute.RecordDatum)
	for i := range a {
		a[i].Release()
		a[i] = nil
	}
	return r.Value
}

type Window interface {
	Init(arrow.Record) bool
	Schema() *arrow.Schema
	Call(map[string]arrow.Array, *array.RecordBuilder, int64, int64)
	Next() (from, to int64, ok bool)
}

func Transform(r arrow.Record, w Window) arrow.Record {
	if !w.Init(r) {
		return r
	}
	columns := r.Columns()
	m := make(map[string]arrow.Array)
	for i := range columns {
		m[r.ColumnName(i)] = columns[i]
	}
	schema := w.Schema()
	build := array.NewRecordBuilder(entry.Pool, schema)
	defer build.Release()
	for lo, hi, ok := w.Next(); ok; lo, hi, ok = w.Next() {
		w.Call(m, build, lo, hi)
	}
	return build.NewRecord()
}

type window struct {
	fields []arrow.Field
	form   func(string, map[string]arrow.Array, array.Builder, int64, int64)
	timestampChunk
}

var _ Window = (*window)(nil)

func (w *window) Schema() *arrow.Schema {
	return arrow.NewSchema(append([]arrow.Field{
		entry.Fields()[entry.Index["timestamp"]],
	}, w.fields...), nil)
}

func (w *window) Init(r arrow.Record) bool {
	for i := 0; i < int(r.NumCols()); i++ {
		if r.ColumnName(i) == "timestamp" {
			a := r.Column(i).(*array.Timestamp).TimestampValues()
			w.timestampChunk = timestampChunk{
				size: int64(len(a)),
				a:    a,
			}
			return true
		}
	}
	return false
}

func (w *window) Call(m map[string]arrow.Array, b *array.RecordBuilder, lo, hi int64) {
	b.Field(0).(*array.TimestampBuilder).Append(w.last)
	for i := range w.fields {
		w.form(w.fields[i].Name, m, b.Field(i+1), lo, hi)
	}
}

type timestampChunk struct {
	start, pos, size int64
	last             arrow.Timestamp
	a                []arrow.Timestamp
}

func (ts *timestampChunk) Next() (lo, hi int64, ok bool) {
	if ts.pos >= ts.size {
		return
	}
	for ; ts.pos < ts.size; ts.pos++ {
		if ts.a[ts.pos] != ts.last {
			lo = ts.start
			hi = ts.pos
			ok = true
			ts.start = ts.pos
			return
		}
	}
	if ts.start < ts.pos && ts.pos <= ts.size {
		lo = ts.start
		hi = ts.pos
		ok = true
		ts.start = ts.pos
	}
	return
}
