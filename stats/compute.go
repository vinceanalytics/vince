package stats

import (
	"context"
	"sort"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/apache/arrow/go/v15/arrow/math"
	"github.com/vinceanalytics/vince/columns"
	"github.com/vinceanalytics/vince/filters"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/logger"
)

type Compute struct {
	mapping map[string]arrow.Array
	view    *float64
	visit   *float64
}

func (c *Compute) Reset(r arrow.Record) {
	clear(c.mapping)
	for i := 0; i < int(r.NumCols()); i++ {
		c.mapping[r.ColumnName(i)] = r.Column(i)
	}
	c.view = nil
	c.visit = nil
}

func (c *Compute) Release() {
	for _, a := range c.mapping {
		a.Release()
	}
	clear(c.mapping)
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
	case v1.Metric_events:
		return c.Events(), nil
	default:
		logger.Fail("Unexpected metric", "err", m)
		return 0, nil
	}
}

func (c *Compute) Events() float64 {
	return float64(c.mapping[columns.Event].Len())
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
	value := math.Float64.Sum(c.mapping[columns.Duration].(*array.Float64))
	return float64(value)
}

func (c *Compute) Bounce() float64 {
	value := math.Int64.Sum(c.mapping[columns.Bounce].(*array.Int64))
	return float64(value)
}

func (c *Compute) Visitors(ctx context.Context) (float64, error) {
	a, err := compute.UniqueArray(ctx, c.mapping[columns.ID])
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
	visit := float64(CalVisits(c.mapping[columns.Session].(*array.Boolean)))
	c.visit = &visit
	return visit
}

func (c *Compute) PageView() float64 {
	if c.view != nil {
		return *c.view
	}
	view := calcPageViews(c.mapping[columns.Event])
	c.view = &view
	return view
}

const viewStr = "pageview"

func calcPageViews(a arrow.Array) (n float64) {
	d := a.(*array.Dictionary)
	x := d.Dictionary().(*array.String)
	for i := 0; i < d.Len(); i++ {
		if d.IsNull(i) {
			continue
		}
		if x.Value(d.GetValueIndex(i)) == viewStr {
			n++
		}
	}
	return
}

func metricsToProjection(f *v1.Filters, me []v1.Metric, props ...v1.Property) []string {
	m := make(map[v1.Filters_Projection]struct{})
	m[v1.Filters_Timestamp] = struct{}{}
	for _, p := range props {
		m[filters.Projection(p)] = struct{}{}
	}
	for _, v := range me {
		switch v {
		case v1.Metric_pageviews:
			m[v1.Filters_Event] = struct{}{}
		case v1.Metric_visitors:
			m[v1.Filters_ID] = struct{}{}
		case v1.Metric_visits:
			m[v1.Filters_Session] = struct{}{}
		case v1.Metric_bounce_rate:
			m[v1.Filters_Session] = struct{}{}
			m[v1.Filters_Bounce] = struct{}{}
		case v1.Metric_visit_duration:
			m[v1.Filters_Duration] = struct{}{}
		case v1.Metric_views_per_visit:
			m[v1.Filters_Event] = struct{}{}
			m[v1.Filters_Duration] = struct{}{}
		}
	}
	cols := make([]string, 0, len(m))

	for k := range m {
		f.Projection = append(f.Projection, k)
		cols = append(cols, filters.ToColumn(k))
	}
	sort.Strings(cols)
	return cols
}
