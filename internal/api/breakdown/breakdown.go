package breakdown

import (
	"cmp"
	"context"
	"math"
	"runtime/trace"
	"slices"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/api/aggregates"
	"github.com/vinceanalytics/vince/internal/fieldset"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/web/query"
)

const (
	visitors  = "visitors"
	visits    = "visits"
	pageviews = "pageviews"
)

type Stats = aggregates.Stats

type Result struct {
	Results []map[string]any `json:"results"`
}

func BreakdownGoals(ctx context.Context, ts *timeseries.Timeseries, site *v1.Site, params *query.Query, metrics []string) (*Result, error) {
	ctx, task := trace.NewTask(ctx, "store.BreakdownGoals")
	defer task.End()
	var (
		pageGoals []string
	)
	efs := query.Filter{
		Op:  "is",
		Key: "name",
	}
	for _, g := range site.Goals {
		if g.Name != "" {
			efs.Value = append(efs.Value, g.Name)
		} else {
			pageGoals = append(pageGoals, g.Path)
		}
	}
	events := new(Result)
	var err error
	if len(efs.Value) > 0 {
		events, err = Breakdown(ctx, ts, site.Domain, params.With(&efs), metrics, models.Field_event)
		if err != nil {
			return nil, err
		}
	}
	pages := new(Result)
	if len(pageGoals) > 0 {
		efs := query.Filter{
			Op:    "is",
			Key:   "page",
			Value: pageGoals,
		}
		pages, err = Breakdown(ctx, ts, site.Domain, params.With(&efs), metrics, models.Field_page)
		if err != nil {
			return nil, err
		}
		for i := range pages.Results {
			m := pages.Results[i]
			m["name"] = m[models.Field_page.String()]
			delete(m, models.Field_page.String())
		}
	}
	result := events
	result.Results = append(result.Results, pages.Results...)
	sortMap(result.Results, "visitors")
	return result, nil
}

func Breakdown(ctx context.Context, ts *timeseries.Timeseries, domain string, params *query.Query, metrics []string, field models.Field) (*Result, error) {
	ctx, task := trace.NewTask(ctx, "store.Breakdown")
	defer task.End()
	return breakdown(
		ctx,
		ts,
		findString(field),
		domain, params, metrics, field, func(property string, values map[string]*Stats) *Result {
			a := &Result{
				Results: make([]map[string]any, 0, len(values)),
			}
			reduce := make([]func(*Stats) float64, len(metrics))
			for i := range metrics {
				reduce[i] = aggregates.StatToValue(metrics[i])
			}
			for k, v := range values {
				v.Compute()
				x := map[string]any{
					property: k,
				}
				for i := range metrics {
					x[metrics[i]] = reduce[i](v)
				}
				a.Results = append(a.Results, x)
			}
			return a
		})
}

func BreakdownExitPages(ctx context.Context, ts *timeseries.Timeseries, domain string, params *query.Query) (*Result, error) {
	ctx, task := trace.NewTask(ctx, "store.BreakdownExitPages")
	defer task.End()
	return breakdown(
		ctx,
		ts,
		findString(models.Field_exit_page),
		domain, params, []string{visitors, visits, pageviews}, models.Field_exit_page, func(property string, values map[string]*Stats) *Result {
			a := &Result{
				Results: make([]map[string]any, 0, len(values)),
			}

			totalPageView := float64(0)
			for _, m := range values {
				m.Compute()
				totalPageView += m.PageViews
			}
			for k, b := range values {
				visits := b.Visits
				visitors := b.Visitors
				exitRate := float64(0)
				if totalPageView != 0 {
					exitRate = math.Floor(visits / totalPageView * 100)
				}
				a.Results = append(a.Results, map[string]any{
					"name":      k,
					"visits":    visits,
					"visitors":  visitors,
					"exit_rate": exitRate,
				})
			}
			return a
		})

}

func BreakdownCity(ctx context.Context, ts *timeseries.Timeseries, domain string, params *query.Query, metrics []string) (*Result, error) {
	ctx, task := trace.NewTask(ctx, "store.BreakdownCity")
	defer task.End()
	return breakdown(ctx, ts,
		findCity,
		domain, params, metrics, models.Field_city, func(property string, values map[uint32]*Stats) *Result {
			a := &Result{
				Results: make([]map[string]any, 0, len(values)),
			}
			reduce := aggregates.Reduce(metrics)
			for code, b := range values {
				b.Compute()
				city := location.GetCity(code)
				value := map[string]any{
					"code": code,
					"name": city.Name,
					"flag": city.Flag,
				}
				reduce(b, value)
				a.Results = append(a.Results, value)
			}
			return a
		})

}

func BreakdownVisitorsWithPercentage(ctx context.Context, ts *timeseries.Timeseries, domain string, params *query.Query, field models.Field) (*Result, error) {
	ctx, task := trace.NewTask(ctx, "store.BreakdownVisitorsWithPercentage")
	defer task.End()
	return breakdown(ctx, ts,
		findString(field),
		domain, params, []string{visitors}, field, func(property string, values map[string]*Stats) *Result {
			a := &Result{
				Results: make([]map[string]any, 0, len(values)),
			}

			var total float64
			for _, m := range values {
				m.Compute()
				total += m.Visitors
			}
			for prop, b := range values {
				vs := b.Visitors
				p := float64(0)
				if total != 0 {
					p = (vs / total) * 100.0
				}
				a.Results = append(a.Results, map[string]any{
					property:     prop,
					visitors:     vs,
					"percentage": math.Floor(p),
				})
			}
			return a
		})

}

func findString(field models.Field) func(context.Context, *timeseries.Timeseries, uint64) string {
	return func(ctx context.Context, tx *timeseries.Timeseries, u uint64) string {
		return tx.Find(ctx, field, u)
	}
}

func findCity(_ context.Context, _ *timeseries.Timeseries, id uint64) uint32 {
	return uint32(id)
}

func breakdown[T cmp.Ordered](ctx context.Context, ts *timeseries.Timeseries, tr func(ctx context.Context, tx *timeseries.Timeseries, id uint64) T, domain string, params *query.Query, metrics []string, field models.Field,
	fn func(property string, values map[T]*Stats) *Result) (*Result, error) {
	values := make(map[T]*Stats)
	fields := fieldset.From(metrics...)
	err := ts.Select(ctx, domain, params.Start(), params.End(), params.Interval(), params.Filter(), func(shard, view uint64, columns *roaring.Bitmap) error {
		all := ts.NewBitmap(ctx, shard, view, field)
		return all.ExtractMutex(columns, func(row uint64, m *roaring.Bitmap) error {
			key := tr(ctx, ts, row)
			sx, ok := values[key]
			if !ok {
				sx = aggregates.NewStats(fields)
				values[key] = sx
			}
			return sx.Read(ctx, ts, shard, view, m, fields)
		})

	})

	if err != nil {
		return nil, err
	}
	a := fn(field.String(), values)
	sortMap(a.Results, visitors)
	return a, nil
}

func sortMap(ls []map[string]any, key string) {
	slices.SortFunc(ls, func(a, b map[string]any) int {
		return cmp.Compare(b[key].(float64), a[key].(float64))
	})
}
