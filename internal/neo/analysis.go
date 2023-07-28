package neo

import (
	"context"
	"sort"

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
	Select() []string
	Filters() []*blocks.Filter
	Analyze(context.Context, arrow.Record)
}

var _ Analysis = (*Base)(nil)

func baseColumnsIndices(f ...string) []int {
	names := baseColumns(f...)
	o := make([]int, 0, len(names))
	for _, k := range names {
		o = append(o, entry.Index[k])
	}
	sort.Ints(o)
	return o
}

func baseColumns(f ...string) []string {
	f = append(f,
		"bounce",
		"duration",
		"id",
		"name",
		"timestamp",
	)
	m := make(map[string]struct{})
	for i := range f {
		m[f[i]] = struct{}{}
	}
	o := make([]string, 0, len(m))
	for k := range m {
		o = append(o, k)
	}
	sort.Strings(o)
	return o
}

type Base struct {
	records []arrow.Record
	columns []string
	indices []int
	filters []*blocks.Filter
}

func NewBase(pick []string, filters ...*blocks.Filter) *Base {
	cols := make([]string, 0, len(pick)+len(filters))
	cols = append(cols, pick...)
	for _, f := range filters {
		cols = append(cols, f.Column)
	}
	names := baseColumns(cols...)
	return &Base{
		columns: names,
		indices: baseColumnsIndices(names...),
		filters: filters,
	}
}

func (b *Base) ColumnIndices() []int {
	return b.indices
}

func (b *Base) Select() []string {
	return b.columns
}

func (b *Base) Filters() []*blocks.Filter {
	return b.filters
}

func (b *Base) Analyze(ctx context.Context, r arrow.Record) {
	b.records = append(b.records,
		selection(ctx, r, b),
	)
}

func selection(ctx context.Context, r arrow.Record, b Analysis) arrow.Record {
	columns := make(map[string]arrow.Array)
	for i := 0; i < int(r.NumCols()); i++ {
		columns[r.ColumnName(i)] = r.Column(i)
	}
	var activeFilter arrow.Array
	for _, f := range b.Filters() {
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
	pick := b.Select()
	f := make([]arrow.Field, 0, len(pick))
	cols := make([]arrow.Array, 0, len(pick))
	sort.Strings(pick)
	for _, n := range pick {
		f = append(f, arrow.Field{
			Name: n,
			Type: columns[n].DataType(),
		})
		cols = append(cols, columns[n])
	}
	record := array.NewRecord(arrow.NewSchema(f, nil), cols, r.NumRows())
	r.Release()
	if activeFilter != nil {
		o := filter(ctx,
			compute.NewDatum(record),
			compute.NewDatum(activeFilter),
		)
		activeFilter.Release()
		return o
	}
	return record
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
