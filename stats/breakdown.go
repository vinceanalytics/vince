package stats

import (
	"context"
	"slices"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/apache/arrow/go/v15/arrow/math"
	"github.com/vinceanalytics/vince/filters"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/logger"
	"github.com/vinceanalytics/vince/session"
)

func BreakDown(ctx context.Context, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error) {
	period := req.Period
	if period == nil {
		period = &v1.TimePeriod{
			Value: &v1.TimePeriod_Base_{
				Base: v1.TimePeriod__30d,
			},
		}
	}
	filter := &v1.Filters{
		List: append(req.Filters, &v1.Filter{
			Property: v1.Property_domain,
			Op:       v1.Filter_equal,
			Value:    req.SiteId,
		}),
	}
	slices.Sort(req.Metrics)
	slices.Sort(req.Property)
	metricsToProjection(filter, req.Metrics, req.Property...)
	from, to := PeriodToRange(time.Now, period)
	r, err := session.Get(ctx).Scan(ctx, from.UnixMilli(), to.UnixMilli(), filter)
	if err != nil {
		logger.Get(ctx).Error("Failed scanning", "err", err)
		return nil, InternalError
	}
	defer r.Release()
	mapping := map[string]arrow.Array{}
	for i := 0; i < int(r.NumCols()); i++ {
		mapping[r.ColumnName(i)] = r.Column(i)
	}
	defer clear(mapping)
	// build key mapping
	b := array.NewUint32Builder(compute.GetAllocator(ctx))
	defer b.Release()
	var result []*v1.BreakDown_Response_Result
	// TODO: run this concurrently
	for _, prop := range req.Property {
		var groups []*v1.BreakDown_Response_Group
		for key, bitmap := range hashProp(mapping[filters.PropToProjection[prop].String()]) {
			b.AppendValues(bitmap.ToArray(), nil)
			idx := b.NewUint32Array()
			var values []*v1.Value
			var visits *float64
			var view *float64
			for _, metric := range req.Metrics {
				var value float64
				switch metric {
				case v1.Metric_pageviews:
					a, err := take(ctx, metric, v1.Filters_Event, mapping, idx)
					if err != nil {
						return nil, err
					}
					count := calcPageViews(a)
					a.Release()
					view = &count
					value = count
				case v1.Metric_visitors:
					a, err := take(ctx, metric, v1.Filters_ID, mapping, idx)
					if err != nil {
						return nil, err
					}
					u, err := compute.Unique(ctx, compute.NewDatumWithoutOwning(a))
					if err != nil {
						a.Release()
						logger.Get(ctx).Error("Failed calculating visitors", "err", err)
						return nil, InternalError
					}
					a.Release()
					value = float64(u.Len())
					u.Release()
				case v1.Metric_visits:
					a, err := take(ctx, metric, v1.Filters_Session, mapping, idx)
					if err != nil {
						return nil, err
					}
					sum := float64(math.Int64.Sum(a.(*array.Int64)))
					a.Release()
					visits = &sum
					value = sum
				case v1.Metric_bounce_rate:
					var vis float64
					if visits != nil {
						vis = *visits
					} else {
						a, err := take(ctx, metric, v1.Filters_Session, mapping, idx)
						if err != nil {
							return nil, err
						}
						vis = float64(math.Int64.Sum(a.(*array.Int64)))
						a.Release()
					}
					a, err := take(ctx, metric, v1.Filters_Bounce, mapping, idx)
					if err != nil {
						return nil, err
					}
					sum := float64(math.Int64.Sum(a.(*array.Int64)))
					a.Release()
					if vis != 0 {
						sum /= vis
					}
					value = sum
				case v1.Metric_visit_duration:
					a, err := take(ctx, metric, v1.Filters_Duration, mapping, idx)
					if err != nil {
						return nil, err
					}
					sum := math.Float64.Sum(a.(*array.Float64))
					a.Release()
					count := float64(a.Len())
					var avg float64
					if count != 0 {
						avg = sum / count
					}
					value = avg
				case v1.Metric_views_per_visit:
					var vis float64
					if visits != nil {
						vis = *visits
					} else {
						a, err := take(ctx, metric, v1.Filters_Session, mapping, idx)
						if err != nil {
							return nil, err
						}
						vis = float64(math.Int64.Sum(a.(*array.Int64)))
						a.Release()
					}
					var vw float64
					if view != nil {
						vw = *view
					} else {
						a, err := take(ctx, metric, v1.Filters_Event, mapping, idx)
						if err != nil {
							return nil, err
						}
						vw = calcPageViews(a)
						a.Release()
					}
					if vis != 0 {
						vw /= vis
					}
					value = vw
				case v1.Metric_events:
					value = float64(idx.Len())
				}
				values = append(values, &v1.Value{
					Metric: metric,
					Value:  value,
				})
			}
			groups = append(groups, &v1.BreakDown_Response_Group{
				Key:    key,
				Values: values,
			})
			idx.Release()
		}
		result = append(result, &v1.BreakDown_Response_Result{
			Property: prop,
			Groups:   groups,
		})
	}
	return &v1.BreakDown_Response{Results: result}, nil
}

func take(ctx context.Context, metric v1.Metric, f v1.Filters_Projection, mapping map[string]arrow.Array, idx *array.Uint32) (arrow.Array, error) {
	a, err := compute.TakeArray(ctx,
		mapping[f.String()], idx,
	)
	if err != nil {
		idx.Release()
		logger.Get(ctx).Error("Failed taking array values",
			"err", err, "metric", metric, "projection", f)
		return nil, InternalError
	}
	return a, nil
}
func hashProp(a arrow.Array) map[string]*roaring.Bitmap {
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
