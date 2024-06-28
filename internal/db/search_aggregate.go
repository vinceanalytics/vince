package db

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf/dsl"
	"github.com/gernest/rows"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/defaults"
	"github.com/vinceanalytics/vince/internal/timeutil"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (db *DB) Aggregate(ctx context.Context, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error) {
	defaults.Set(req)
	err := validate.Validate(req)
	if err != nil {
		return nil, err
	}
	m := dupe(req.Metrics)
	a := newAggregate()
	query := &aggregateQuery{a: a.View(m)}
	from, to := periodToRange(req.Period, req.Date)
	err = db.Search(from, to, append(req.Filters, &v1.Filter{
		Property: v1.Property_domain,
		Op:       v1.Filter_equal,
		Value:    req.SiteId,
	}), query)
	if err != nil {
		return nil, err
	}
	res := &v1.Aggregate_Response{
		Results: make(map[string]float64),
	}
	for _, mx := range m {
		res.Results[mx.String()] = a.Result(mx)
	}
	return res, nil
}

func dupe[T cmp.Ordered](a []T) []T {
	m := make(map[T]struct{})
	for i := range a {
		m[a[i]] = struct{}{}
	}
	o := make([]T, 0, len(m))
	for k := range m {
		o = append(o, k)
	}
	slices.Sort(o)
	return o
}

type aggregateQuery struct {
	a applyList
}

var _ Query = (*aggregateQuery)(nil)

func (a *aggregateQuery) View(_ time.Time) View {
	return a.a
}

type aggregate struct {
	visitors    roaring64.Bitmap
	visits      roaring64.Bitmap
	views       roaring64.Bitmap
	bounceTrue  roaring64.Bitmap
	bounceFalse roaring64.Bitmap
	events      roaring64.Bitmap
	duration    roaring64.BSI

	cache applyList
}

func newAggregate() *aggregate {
	return &aggregate{duration: *roaring64.NewDefaultBSI()}
}

func (a *aggregate) Result(m v1.Metric) float64 {
	switch m {
	case v1.Metric_visitors:
		return float64(a.Visitors())
	case v1.Metric_visits:
		return float64(a.Visits())
	case v1.Metric_pageviews:
		return float64(a.Views())
	case v1.Metric_views_per_visit:
		return a.ViewsPerVisit()
	case v1.Metric_bounce_rate:
		return a.BounceRate()
	case v1.Metric_visit_duration:
		d := a.Duration()
		// convert to seconds
		return (time.Duration(d) * time.Millisecond).Seconds()
	case v1.Metric_events:
		return float64(a.Events())
	default:
		return 0
	}
}

func (a *aggregate) View(m []v1.Metric) applyList {
	if len(a.cache) > 0 {
		return a.cache
	}
	o := make(map[string]struct{})
	for i := range m {
		switch m[i] {
		case v1.Metric_visitors:
			o["visitors"] = struct{}{}
		case v1.Metric_visits:
			o["visits"] = struct{}{}
		case v1.Metric_pageviews:
			o["views"] = struct{}{}
		case v1.Metric_views_per_visit:
			o["views"] = struct{}{}
			o["visits"] = struct{}{}
		case v1.Metric_bounce_rate:
			o["bounce"] = struct{}{}
			o["visits"] = struct{}{}
		case v1.Metric_visit_duration:
			o["duration"] = struct{}{}
		case v1.Metric_events:
			o["events"] = struct{}{}
		}
	}
	ls := make(applyList, 0, len(o))
	for k := range o {
		switch k {
		case "visitors":
			ls = append(ls, a.applyVisitors)
		case "visits":
			ls = append(ls, a.applyVisits)
		case "views":
			ls = append(ls, a.applyViews)
		case "bounce":
			ls = append(ls, a.applyBounce)
		case "duration":
			ls = append(ls, a.applyDuration)
		case "events":
			ls = append(ls, a.applyEvents)
		}
	}
	a.cache = ls
	return ls
}

type applyList []func(*Tx, *rows.Row) error

var _ View = (*applyList)(nil)

func (ls applyList) Apply(tx *Tx, columns *rows.Row) error {
	for i := range ls {
		err := ls[i](tx, columns)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *aggregate) Reset() {
	a.visitors.Clear()
	a.visits.Clear()
	a.views.Clear()
	a.events.Clear()
	a.bounceTrue.Clear()
	a.bounceFalse.Clear()
	a.duration.ClearValues(a.duration.GetExistenceBitmap())
}

func (a *aggregate) BounceRate() float64 {
	bounce := a.Bounce()
	visits := a.Visits()
	if visits != 0 {
		return float64(bounce) / float64(visits)
	}
	return 0
}

func (a *aggregate) ViewsPerVisit() float64 {
	views := a.Views()
	visits := a.Visits()
	if visits != 0 {
		return float64(views) / float64(visits)
	}
	return 0
}

func (a *aggregate) Events() uint64 {
	return a.events.GetCardinality()
}

func (a *aggregate) Visitors() uint64 {
	return a.visitors.GetCardinality()
}

func (a *aggregate) Visits() uint64 {
	return a.visits.GetCardinality()
}

func (a *aggregate) Views() uint64 {
	return a.views.GetCardinality()
}

func (a *aggregate) Bounce() uint64 {
	yes := a.bounceTrue.GetCardinality()
	no := a.bounceFalse.GetCardinality()
	if no < yes {
		return yes - no
	}
	return 0
}

func (a *aggregate) Duration() uint64 {
	b := a.duration.GetExistenceBitmap()
	sum, _ := a.duration.Sum(b)
	return uint64(sum)
}

func (a *aggregate) applyEvents(tx *Tx, columns *rows.Row) error {
	return columns.RangeColumns(func(u uint64) error {
		a.events.Add(u)
		return nil
	})
}
func (a *aggregate) applyVisitors(tx *Tx, columns *rows.Row) error {
	view := new(ViewFmt).Format(tx.View, "uid")
	add := func(_, value uint64) error {
		a.visitors.Add(value)
		return nil
	}
	return dsl.ExtractValuesBSI(tx.Tx, view, tx.Shard, columns, add)
}

func (a *aggregate) applyDuration(tx *Tx, columns *rows.Row) error {
	view := new(ViewFmt).Format(tx.View, "duration")
	add := func(column, value uint64) error {
		a.duration.SetValue(column, int64(value))
		return nil
	}
	return dsl.ExtractValuesBSI(tx.Tx, view, tx.Shard, columns, add)
}

func (a *aggregate) applyVisits(tx *Tx, columns *rows.Row) error {
	return a.true(&a.visits, "session", tx, columns)
}

func (a *aggregate) applyViews(tx *Tx, columns *rows.Row) error {
	return a.true(&a.visits, "views", tx, columns)
}

func (a *aggregate) applyBounce(tx *Tx, columns *rows.Row) error {
	return errors.Join(
		a.true(&a.bounceTrue, "bounce", tx, columns),
		a.false(&a.bounceFalse, "bounce", tx, columns),
	)
}

func (a *aggregate) true(o *roaring64.Bitmap, field string, tx *Tx, columns *rows.Row) error {
	return a.boolean(o, field, true, tx, columns)
}

func (a *aggregate) false(o *roaring64.Bitmap, field string, tx *Tx, columns *rows.Row) error {
	return a.boolean(o, field, false, tx, columns)
}

func (a *aggregate) boolean(o *roaring64.Bitmap, field string, cond bool, tx *Tx, columns *rows.Row) error {
	view := new(ViewFmt).Format(tx.View, field)
	var r *rows.Row
	var err error
	if cond {
		r, err = dsl.True(tx.Tx, view, tx.Shard, columns)
	} else {
		r, err = dsl.False(tx.Tx, view, tx.Shard, columns)
	}
	if err != nil {
		return err
	}
	return r.RangeColumns(func(u uint64) error {
		o.Add(u)
		return nil
	})
}

func periodToRange(period *v1.TimePeriod, tsDate *timestamppb.Timestamp) (start, end time.Time) {
	date := tsDate.AsTime()
	switch e := period.Value.(type) {
	case *v1.TimePeriod_Base_:
		switch e.Base {
		case v1.TimePeriod_day:
			end = date
			start = timeutil.BeginDay(end)
		case v1.TimePeriod__7d:
			end = date
			start = end.AddDate(0, 0, -6)
		case v1.TimePeriod__30d:
			end = date
			start = end.AddDate(0, 0, -30)
		case v1.TimePeriod_mo:
			end = date
			start = timeutil.BeginMonth(end)
			end = timeutil.EndMonth(end)
		case v1.TimePeriod__6mo:
			end = timeutil.EndMonth(date)
			start = timeutil.BeginMonth(end.AddDate(0, -5, 0))
		case v1.TimePeriod__12mo:
			end = timeutil.EndMonth(date)
			start = timeutil.BeginMonth(end.AddDate(0, -11, 0))
		case v1.TimePeriod_year:
			end = timeutil.EndYear(date)
			start = timeutil.BeginYear(end)
		}

	case *v1.TimePeriod_Custom_:
		end = e.Custom.End.AsTime()
		start = e.Custom.Start.AsTime()
	}
	return
}

type nowFunc func() time.Time

func parseDate(now nowFunc, ts *timestamppb.Timestamp) time.Time {
	if ts != nil {
		return ts.AsTime()
	}
	return timeutil.EndDay(now())
}

func ValidByPeriod(period *v1.TimePeriod, i v1.Interval) bool {
	switch e := period.Value.(type) {
	case *v1.TimePeriod_Base_:
		switch e.Base {
		case v1.TimePeriod_day:
			return i == v1.Interval_minute || i == v1.Interval_hour
		case v1.TimePeriod__7d:
			return i == v1.Interval_hour || i == v1.Interval_date
		case v1.TimePeriod_mo, v1.TimePeriod__30d:
			return i == v1.Interval_date || i == v1.Interval_week
		case v1.TimePeriod__6mo, v1.TimePeriod__12mo, v1.TimePeriod_year:
			return i == v1.Interval_date || i == v1.Interval_week || i == v1.Interval_month
		default:
			return false
		}
	case *v1.TimePeriod_Custom_:
		return i == v1.Interval_date || i == v1.Interval_week || i == v1.Interval_month
	default:
		return false
	}
}
