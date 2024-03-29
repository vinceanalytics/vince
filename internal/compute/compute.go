package compute

import (
	"cmp"
	"context"
	"sort"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/bitutil"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/apache/arrow/go/v15/arrow/math"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/columns"
	"github.com/vinceanalytics/vince/internal/filters"
	"github.com/vinceanalytics/vince/internal/logger"
)

type Compute struct {
	Mapping map[string]arrow.Array
	view    *float64
	visit   *float64
}

func NewCompute(r arrow.Record) *Compute {
	m := make(map[string]arrow.Array)
	for i := 0; i < int(r.NumCols()); i++ {
		m[r.ColumnName(i)] = r.Column(i)
	}
	return &Compute{Mapping: m}
}

func (c *Compute) Reset(r arrow.Record) {
	clear(c.Mapping)
	if r != nil {
		for i := 0; i < int(r.NumCols()); i++ {
			c.Mapping[r.ColumnName(i)] = r.Column(i)
		}
	}

	c.view = nil
	c.visit = nil
}

func (c *Compute) Release() {
	for _, a := range c.Mapping {
		a.Release()
	}
	clear(c.Mapping)
	c.view = nil
	c.visit = nil
}

func (c *Compute) Metric(ctx context.Context, m v1.Metric) (float64, error) {
	switch m {
	case v1.Metric_pageviews:
		return c.PageView(), nil
	case v1.Metric_visitors:
		return c.Visitors(ctx)
	case v1.Metric_visits:
		return c.Visits(), nil
	case v1.Metric_bounce_rate:
		return c.BounceRate(), nil
	case v1.Metric_visit_duration:
		return c.VisitDuration(), nil
	case v1.Metric_views_per_visit:
		return c.ViewsPerVisits(), nil
	case v1.Metric_events:
		return c.Events(), nil
	default:
		logger.Fail("Unexpected metric", "err", m)
		return 0, nil
	}
}

func (c *Compute) Events() float64 {
	return float64(c.Mapping[columns.Event].Len())
}

func (c *Compute) ViewsPerVisits() float64 {
	views := c.PageView()
	visits := c.Visits()
	if visits != 0 {
		return views / visits
	}
	return 0
}

func (c *Compute) VisitDuration() float64 {
	duration := c.Duration()
	visits := c.Visits()
	if visits != 0 {
		return duration / visits
	}
	return 0
}

func (c *Compute) BounceRate() float64 {
	bounce := c.Bounce()
	visits := c.Visits()
	if visits != 0 {
		return bounce / visits
	}
	return 0
}

func (c *Compute) Duration() float64 {
	value := math.Float64.Sum(c.Mapping[columns.Duration].(*array.Float64))
	return float64(value)
}

func (c *Compute) Bounce() float64 {
	return float64(CalcBounce(c.Mapping[columns.Bounce].(*array.Boolean)))
}

func (c *Compute) Visitors(ctx context.Context) (float64, error) {
	a, err := compute.UniqueArray(ctx, c.Mapping[columns.ID])
	if err != nil {
		return 0, err
	}
	defer a.Release()
	return float64(a.Len()), nil
}

func (c *Compute) Visits() float64 {
	if c.visit != nil {
		return *c.visit
	}
	visit := float64(CalVisits(c.Mapping[columns.Session].(*array.Boolean)))
	c.visit = &visit
	return visit
}

func (c *Compute) PageView() float64 {
	if c.view != nil {
		return *c.view
	}
	view := float64(countSetBits(c.Mapping[columns.View].(*array.Boolean)))
	c.view = &view
	return view
}

func MetricsToProjection(f *v1.Filters, me []v1.Metric, props ...v1.Property) []string {
	m := make(map[v1.Filters_Projection]struct{})
	m[v1.Filters_timestamp] = struct{}{}
	for _, p := range props {
		m[filters.Projection(p)] = struct{}{}
	}
	for _, v := range me {
		switch v {
		case v1.Metric_pageviews:
			m[v1.Filters_view] = struct{}{}
		case v1.Metric_visitors:
			m[v1.Filters_id] = struct{}{}
		case v1.Metric_visits:
			m[v1.Filters_session] = struct{}{}
		case v1.Metric_bounce_rate:
			m[v1.Filters_session] = struct{}{}
			m[v1.Filters_bounce] = struct{}{}
		case v1.Metric_visit_duration:
			m[v1.Filters_session] = struct{}{}
			m[v1.Filters_duration] = struct{}{}
		case v1.Metric_views_per_visit:
			m[v1.Filters_view] = struct{}{}
			m[v1.Filters_duration] = struct{}{}
		case v1.Metric_events:
			m[v1.Filters_event] = struct{}{}
		}
	}
	cols := make([]string, 0, len(m))

	for k := range m {
		f.Projection = append(f.Projection, k)
		cols = append(cols, k.String())
	}
	sort.Strings(cols)
	return cols
}

// We store sessions as boolean. True for new sessions and false otherwise.
// Visits is the same as the number of set bits.
func CalVisits(a *array.Boolean) int {
	return countSetBits(a)
}

func countSetBits(a *array.Boolean) int {
	vals := a.Data().Buffers()[1]
	if vals != nil {
		return bitutil.CountSetBits(vals.Bytes(), 0, a.Len())
	}
	return 0
}

func CalcBounce(a *array.Boolean) int {
	nulls := a.NullN()
	set := countSetBits(a)
	switch cmp.Compare(set, nulls) {
	case -1:
		return 0
	case 1:
		return set - nulls
	default:
		return 0
	}
}

func Take(ctx context.Context, a arrow.Array, idx *array.Uint32) (arrow.Array, error) {
	if a.DataType().ID() != arrow.DICTIONARY {
		return compute.TakeArray(ctx, a, idx)
	}
	x := a.(*array.Dictionary)
	xv := x.Dictionary().(*array.String)
	o := array.NewDictionaryBuilder(
		compute.GetAllocator(ctx),
		a.DataType().(*arrow.DictionaryType),
	).(*array.BinaryDictionaryBuilder)
	defer o.Release()
	for _, i := range idx.Uint32Values() {
		if x.IsNull(int(i)) {
			o.AppendNull()
			continue
		}
		o.Append([]byte(xv.Value(x.GetValueIndex(int(i)))))
	}
	return o.NewArray(), nil
}

func HashProp(a arrow.Array) map[string]*roaring.Bitmap {
	o := make(map[string]*roaring.Bitmap)
	d := a.(*array.Dictionary)
	s := d.Dictionary().(*array.String)
	for i := 0; i < a.Len(); i++ {
		if d.IsNull(i) {
			continue
		}
		x := s.Value(d.GetValueIndex(i))
		b, ok := o[x]
		if !ok {
			b = new(roaring.Bitmap)
			o[x] = b
		}
		b.Add(uint32(i))
	}
	return o
}
