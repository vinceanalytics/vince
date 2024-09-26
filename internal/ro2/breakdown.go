package ro2

import (
	"cmp"
	"math"
	"slices"
	"time"

	groar "github.com/gernest/roaring"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/bsi"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/cursor"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/mutex"
	"github.com/vinceanalytics/vince/internal/rbf/quantum"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
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
	return o.breakdown(domain, params, metrics, field, func(property string, values map[string]*roaring64.Bitmap, m *Data) *Result {
		a := &Result{
			Results: make([]map[string]any, 0, len(values)),
		}
		for k, v := range values {
			x := map[string]any{
				property: k,
			}
			for i := range metrics {
				x[metrics[i]] = m.Compute(metrics[i], v)
			}
			a.Results = append(a.Results, x)
		}
		return a
	})
}

func (o *Store) BreakdownExitPages(domain string, params *query.Query) (*Result, error) {
	return o.breakdown(domain, params, []string{visitors, visits, pageviews}, alicia.EXIT_PAGE, func(property string, values map[string]*roaring64.Bitmap, m *Data) *Result {
		a := &Result{
			Results: make([]map[string]any, 0, len(values)),
		}

		totalPageView := float64(m.View(nil))
		for k, b := range values {
			visits := float64(m.Visits(b))
			visitors := float64(m.Visitors(b))
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
	values := make(map[uint32]*roaring64.Bitmap)
	m := NewData()
	defer m.Release()
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
						err = viewCu(rtx, string(property)+string(view), func(rCu *rbf.Cursor) error {
							return bsi.Extract(rCu, shard, dRow, func(column uint64, value int64) {
								code, ok := values[uint32(value)]
								if !ok {
									code = roaring64.New()
									values[uint32(value)] = code
								}
								code.Add(column)
							})
						})
						if err != nil {
							return err
						}
						return m.ReadFields(rtx, string(view), shard, dRow, fields...)
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
		vs := m.Compute(visitors, b)
		city := location.GetCity(code)
		a.Results = append(a.Results, map[string]any{
			visitors: vs,
			"code":   code,
			"name":   city.Name,
			"flag":   city.Flag,
		})
	}
	sortMap(a.Results, visitors)
	return a, nil
}

func (o *Store) BreakdownVisitorsWithPercentage(domain string, params *query.Query, field alicia.Field) (*Result, error) {
	return o.breakdown(domain, params, []string{visitors}, field, func(property string, values map[string]*roaring64.Bitmap, m *Data) *Result {
		a := &Result{
			Results: make([]map[string]any, 0, len(values)),
		}
		total := m.Compute(visitors, nil)
		for prop, b := range values {
			vs := m.Compute(visitors, b)
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
	fn func(property string, values map[string]*roaring64.Bitmap, data *Data) *Result) (*Result, error) {
	m := NewData()
	defer m.Release()
	values := make(map[string]*roaring64.Bitmap)
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
						err = viewCu(rtx, string(property)+string(view), func(rCu *rbf.Cursor) error {
							return mutex.Distinct(rCu, dRow, func(row uint64, columns *groar.Container) error {
								key := tx.Find(uint64(field), row)
								value := roaring64.New()
								groar.ContainerCallback(columns, func(u uint16) {
									value.Add(uint64(u))
								})
								values[key] = value
								return nil
							})
						})
						if err != nil {
							return err
						}
						return m.ReadFields(rtx, string(view), shard, dRow, fields...)
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
	a := fn(property, values, m)
	sortMap(a.Results, visitors)
	return a, nil
}

func sortMap(ls []map[string]any, key string) {
	slices.SortFunc(ls, func(a, b map[string]any) int {
		return cmp.Compare(b[key].(float64), a[key].(float64))
	})
}
