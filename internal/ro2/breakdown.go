package ro2

import (
	"cmp"
	"math"
	"slices"
	"time"

	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/bsi"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/cursor"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/mutex"
	"github.com/vinceanalytics/vince/internal/rbf/quantum"
	"github.com/vinceanalytics/vince/internal/web/query"
	"google.golang.org/protobuf/encoding/protowire"
)

const (
	visitors  = "visitors"
	visits    = "visits"
	pageviews = "pageviews"
)

type Result struct {
	Results []map[string]any `json:"results"`
}

func (o *Store) Breakdown(domain string, params *query.Query, metrics []string, field alicia.Field) (*Result, error) {
	return o.breakdown(domain, params, metrics, field, func(property string, values map[string]*Stats) *Result {
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
	return o.breakdown(domain, params, []string{visitors, visits, pageviews}, alicia.EXIT_PAGE, func(property string, values map[string]*Stats) *Result {
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

func (o *Store) BreakdownCity(domain string, params *query.Query) (*Result, error) {
	values := make(map[uint32]*Stats)
	property := "city"
	fields := MetricsToProject([]string{visitors})
	err := o.View(func(tx *Tx) error {
		domainId, ok := tx.ID(uint64(alicia.DOMAIN), domain)
		if !ok {
			return nil
		}
		shards := o.Shards(tx)
		match := tx.compile(params.Filter())
		for _, shard := range shards {
			err := o.shards.View(shard, func(rtx *rbf.Tx) error {
				f := quantum.NewField()
				defer f.Release()
				fn := f.Day
				switch params.Interval() {
				case query.Minute:
					fn = f.Minute
				case query.Hour:
					fn = f.Hour
				case query.Week:
					fn = f.Week
				case query.Month:
					fn = f.Month
				}
				if params.All() {
					fn = func(name string, start, end time.Time, fn func([]byte) error) error {
						return fn([]byte(name))
					}
				}
				return fn(domainField, params.Start(), params.End(), func(b []byte) error {
					return viewCu(rtx, string(b), func(rCu *rbf.Cursor) error {
						dRow, err := cursor.Row(rCu, shard, domainId)
						if err != nil {
							return err
						}
						if dRow.IsEmpty() {
							return nil
						}
						view := b[len(domainField):]
						dRow, err = match.Apply(rtx, shard, view, dRow)
						if err != nil {
							return err
						}
						if dRow.IsEmpty() {
							return nil
						}
						return viewCu(rtx, string(property)+string(view), func(rCu *rbf.Cursor) error {
							return bsi.Distinct(rCu, shard, dRow, func(value uint64, columns *rows.Row) error {
								code, ok := values[uint32(value)]
								if !ok {
									code = new(Stats)
									values[uint32(value)] = code
								}
								return code.ReadFields(rtx, string(view), shard, columns, fields...)
							})
						})
					})
				})
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	a := &Result{
		Results: make([]map[string]any, 0, len(values)),
	}
	for code, b := range values {
		b.Compute()
		city := location.GetCity(code)
		a.Results = append(a.Results, map[string]any{
			visitors: b.Visitors,
			"code":   code,
			"name":   city.Name,
			"flag":   city.Flag,
		})
	}
	sortMap(a.Results, visitors)
	return a, nil
}

func (o *Store) BreakdownVisitorsWithPercentage(domain string, params *query.Query, field alicia.Field) (*Result, error) {
	return o.breakdown(domain, params, []string{visitors}, field, func(property string, values map[string]*Stats) *Result {
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

func (o *Store) breakdown(domain string, params *query.Query, metrics []string,
	field alicia.Field,
	fn func(property string, values map[string]*Stats) *Result) (*Result, error) {
	values := make(map[string]*Stats)
	property := string(fields.ByNumber(protowire.Number(field)).Name())
	fields := MetricsToProject(metrics)

	err := o.View(func(tx *Tx) error {
		domainId, ok := tx.ID(uint64(alicia.DOMAIN), domain)
		if !ok {
			return nil
		}
		shards := o.Shards(tx)
		match := tx.compile(params.Filter())
		for _, shard := range shards {
			err := o.shards.View(shard, func(rtx *rbf.Tx) error {
				f := quantum.NewField()
				defer f.Release()
				fn := f.Day
				switch params.Interval() {
				case query.Minute:
					fn = f.Minute
				case query.Hour:
					fn = f.Hour
				case query.Week:
					fn = f.Week
				case query.Month:
					fn = f.Month
				}
				if params.All() {
					fn = func(name string, start, end time.Time, fn func([]byte) error) error {
						return fn([]byte(name))
					}
				}
				return fn(domainField, params.Start(), params.End(), func(b []byte) error {
					return viewCu(rtx, string(b), func(rCu *rbf.Cursor) error {
						dRow, err := cursor.Row(rCu, shard, domainId)
						if err != nil {
							return err
						}
						if dRow.IsEmpty() {
							return nil
						}
						view := b[len(domainField):]
						dRow, err = match.Apply(rtx, shard, view, dRow)
						if err != nil {
							return err
						}
						if dRow.IsEmpty() {
							return nil
						}
						return viewCu(rtx, string(property)+string(view), func(rCu *rbf.Cursor) error {
							return mutex.Distinct(rCu, dRow, func(row uint64, columns *rows.Row) error {
								key := tx.Find(uint64(field), row)
								sx, ok := values[key]
								if !ok {
									sx = new(Stats)
									values[key] = sx
								}
								return sx.ReadFields(rtx, string(view), shard, columns, fields...)
							})
						})
					})
				})
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	a := fn(property, values)
	sortMap(a.Results, visitors)
	return a, nil
}

func sortMap(ls []map[string]any, key string) {
	slices.SortFunc(ls, func(a, b map[string]any) int {
		return cmp.Compare(b[key].(float64), a[key].(float64))
	})
}
