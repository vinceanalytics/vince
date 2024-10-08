package ro2

import (
	"cmp"
	"math"
	"slices"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/fieldset"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/web/query"
)

const (
	visitors  = "visitors"
	visits    = "visits"
	pageviews = "pageviews"
)

type Result struct {
	Results []map[string]any `json:"results"`
}

func (o *Store) Breakdown(domain string, params *query.Query, metrics []string, field v1.Field) (*Result, error) {
	return breakdown(
		o,
		findString(field),
		domain, params, metrics, field, func(property string, values map[string]*Stats) *Result {
			a := &Result{
				Results: make([]map[string]any, 0, len(values)),
			}
			reduce := make([]func(*Stats) float64, len(metrics))
			for i := range metrics {
				reduce[i] = StatToValue(metrics[i])
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

func (o *Store) BreakdownExitPages(domain string, params *query.Query) (*Result, error) {
	return breakdown(
		o,
		findString(v1.Field_exit_page),
		domain, params, []string{visitors, visits, pageviews}, v1.Field_exit_page, func(property string, values map[string]*Stats) *Result {
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

func (o *Store) BreakdownCity(domain string, params *query.Query, metrics []string) (*Result, error) {
	return breakdown(o,
		findCity,
		domain, params, metrics, v1.Field_city, func(property string, values map[uint32]*Stats) *Result {
			a := &Result{
				Results: make([]map[string]any, 0, len(values)),
			}
			reduce := Reduce(metrics)
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

func (o *Store) BreakdownVisitorsWithPercentage(domain string, params *query.Query, field v1.Field) (*Result, error) {
	return breakdown(o,
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

func findString(field v1.Field) func(*Tx, uint64) string {
	return func(tx *Tx, u uint64) string {
		return tx.Find(field, u)
	}
}

func findCity(_ *Tx, id uint64) uint32 {
	return uint32(id)
}

func breakdown[T cmp.Ordered](o *Store, tr func(tx *Tx, id uint64) T, domain string, params *query.Query, metrics []string, field v1.Field,
	fn func(property string, values map[T]*Stats) *Result) (*Result, error) {
	values := make(map[T]*Stats)
	fields := fieldset.From(metrics...)

	err := o.View(func(tx *Tx) error {
		return tx.Select(domain, params.Start(), params.End(), params.Interval(), params.Filter(), func(shard, view uint64, columns *roaring.Bitmap) error {
			all, err := tx.TransposeSet(shard, view, field, columns)
			if err != nil {
				return err
			}
			m := roaring.NewBitmap()
			for id, v := range all {
				key := tr(tx, uint64(id))
				sx, ok := values[key]
				if !ok {
					sx = NewStats(fields)
					values[key] = sx
				}
				m.Reset()
				m.SetMany(v)
				err := sx.Read(tx, shard, view, m, fields)
				if err != nil {
					return err
				}
			}
			return nil
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
