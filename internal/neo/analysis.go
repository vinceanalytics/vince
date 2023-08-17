package neo

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/compute"
	"github.com/apache/arrow/go/v13/arrow/math"
	"github.com/apache/arrow/go/v13/arrow/scalar"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/must"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

type Analysis interface {
	ColumnIndices() []int
	Analyze(context.Context, arrow.Record)
}

var _ Analysis = (*Base)(nil)

type Base struct {
	columns []string
	indices []int
	filters []*v1.Filter
	build   *array.RecordBuilder
}

func NewBase(pick []string, filters ...*v1.Filter) *Base {
	b := basePool.Get().(*Base)
	return b.Init(pick, filters...)
}

var basePool = &sync.Pool{
	New: func() any {
		return &Base{
			columns: make([]string, 0, len(entry.Index)),
			indices: make([]int, 0, len(entry.Index)),
			filters: make([]*v1.Filter, 0, len(entry.Index)),
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

func (b *Base) Init(pick []string, filters ...*v1.Filter) *Base {
	b.columns = append(b.columns,
		"bounce",
		"duration",
		"id",
		"name",
		"session",
		"timestamp",
	)

	seen := map[string]struct{}{
		"bounce":    {},
		"duration":  {},
		"id":        {},
		"name":      {},
		"session":   {},
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

func apply(ctx context.Context, f *v1.Filter, a arrow.Array) arrow.Array {
	switch e := f.Value.(type) {
	case *v1.Filter_Str:
		return call(ctx, f.Op.String(),
			compute.NewDatum(a),
			compute.NewDatum(scalar.NewStringScalar(e.Str)),
		)
	case *v1.Filter_Timestamp:
		return call(ctx, f.Op.String(),
			compute.NewDatum(a),
			compute.NewDatum(scalar.NewTimestampScalar(arrow.Timestamp(e.Timestamp),
				&arrow.TimestampType{Unit: arrow.Millisecond})),
		)
	case *v1.Filter_Duration:
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
	o := must.
		Must(compute.CallFunction(ctx, f, nil, a...))("failed applying comparison operator ", f)
	r := o.(*compute.ArrayDatum).MakeArray()
	for i := range a {
		a[i].Release()
		a[i] = nil
	}
	o.Release()
	return r
}

func sum(a arrow.Array) int64 {
	return math.Int64.Sum(a.(*array.Int64))
}

func filter(ctx context.Context, a ...compute.Datum) arrow.Record {
	o := must.
		Must(compute.CallFunction(ctx, "filter",
			compute.DefaultFilterOptions(), a...))("failed applying filter")
	r := o.(*compute.RecordDatum)
	for i := range a {
		a[i].Release()
		a[i] = nil
	}
	return r.Value
}

type Metrics struct {
	Visitors        float64
	PageViews       float64
	Sessions        float64
	BounceRate      float64
	SessionDuration float64
}

func (m *Metrics) Compute(ctx context.Context, r arrow.Record) {
	*m = Metrics{}
	for i := 0; i < int(r.NumCols()); i++ {
		switch r.ColumnName(i) {
		case "id":
			a := must.Must(compute.UniqueArray(ctx, r.Column(i)))(
				"failed computing unique array for id column",
			)
			m.Visitors = float64(a.Len())
			a.Release()
		case "name":
			m.PageViews = float64(r.Column(i).Len())
		case "session":
			m.Sessions = float64(sum(r.Column(i)))
		case "bounce":
			m.BounceRate = float64(sum(r.Column(i)))
		case "duration":
			m.SessionDuration = float64(sum(r.Column(i)))
		}
	}
	if m.Sessions != 0 {
		m.BounceRate = m.BounceRate / m.Sessions * 100
		m.SessionDuration = m.SessionDuration / m.Sessions
	} else {
		m.BounceRate = 0
		m.SessionDuration = 0
	}
}

func (m *Metrics) Record(b *array.StructBuilder) {
	b.Append(true)
	b.FieldBuilder(0).(*array.Float64Builder).Append(m.Visitors)
	b.FieldBuilder(1).(*array.Float64Builder).Append(m.PageViews)
	b.FieldBuilder(2).(*array.Float64Builder).Append(m.Sessions)
	b.FieldBuilder(3).(*array.Float64Builder).Append(m.BounceRate)
	b.FieldBuilder(4).(*array.Float64Builder).Append(m.SessionDuration)
}

var computedFields = []arrow.Field{
	{Name: "visitors", Type: arrow.PrimitiveTypes.Float64},
	{Name: "page_views", Type: arrow.PrimitiveTypes.Float64},
	{Name: "sessions", Type: arrow.PrimitiveTypes.Float64},
	{Name: "bounce_rate", Type: arrow.PrimitiveTypes.Float64},
	{Name: "session_duration", Type: arrow.PrimitiveTypes.Float64},
}

var metricsField = arrow.Field{
	Name: "metrics", Type: arrow.StructOf(computedFields...),
}

func computedPartition(names ...string) *arrow.Schema {
	fields := []arrow.Field{
		entry.Fields()[entry.Index["timestamp"]],
	}
	if len(names) == 0 {
		return arrow.NewSchema(append(fields, metricsField), nil)
	}
	sort.Strings(names)
	for _, f := range names {
		fields = append(fields, arrow.Field{
			Name: f,
			Type: arrow.ListOf(
				arrow.StructOf(
					append([]arrow.Field{
						{
							Name: "value", Type: arrow.BinaryTypes.String,
						},
					}, metricsField)...,
				),
			),
		})
	}
	return arrow.NewSchema(fields, nil)
}

func Transform(ctx context.Context, r arrow.Record, step, truncate time.Duration, partitions ...string) arrow.Record {
	ctx = entry.Context(ctx)
	defer r.Release()
	var ts []arrow.Timestamp
	for i := 0; i < int(r.NumRows()); i++ {
		if r.ColumnName(i) == "timestamp" {
			ts = r.Column(i).(*array.Timestamp).TimestampValues()
			break
		}
	}
	must.Assert(ts != nil)("passed a record for transformation without timestamp")
	hasPartitions := len(partitions) > 0
	b := array.NewRecordBuilder(entry.Pool, computedPartition(partitions...))
	defer b.Release()
	tsField := b.Field(0).(*array.TimestampBuilder)
	var result Metrics
	if !hasPartitions && step == 0 {
		tsField.Append(arrow.Timestamp(core.Now(ctx).UnixMilli()))
		result.Compute(ctx, r)
		result.Record(b.Field(1).(*array.StructBuilder))
		return b.NewRecord()
	}
	bound := func(ts arrow.Timestamp) arrow.Timestamp {
		a := ts.ToTime(arrow.Millisecond).Add(step)
		if truncate != 0 {
			a = a.Truncate(truncate)
		}
		return arrow.Timestamp(a.UnixMilli())
	}
	boundary := bound(ts[0])
	lo := 0
	for i := range ts {
		if ts[i] < boundary {
			continue
		}
		tsField.Append(boundary)
		slice := r.NewSlice(int64(lo), int64(i))
		if !hasPartitions {
			result.Compute(ctx, slice)
			result.Record(b.Field(1).(*array.StructBuilder))
		}
		boundary = bound(ts[i])
		lo = i
		slice.Release()
	}
	return b.NewRecord()
}
