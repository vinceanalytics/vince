package ro2

import (
	"cmp"
	"math"
	"slices"

	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

const (
	visitors  = "visitors"
	visits    = "visits"
	pageviews = "pageviews"
)

type Result struct {
	Results []map[string]any `json:"results"`
}

func (o *Store) Breakdown(start, end int64, domain string, filter Filter, metrics []string, field alicia.Field) (*Result, error) {
	m := NewData()
	defer m.Release()
	values := make(map[string]*roaring64.Bitmap)
	o.Select(start, end, domain, filter, func(tx *Tx, shard uint64, match *roaring64.Bitmap) error {
		tx.ExtractBSI(shard, uint64(field), match, func(row uint64, c int64) {
			value := tx.Find(uint64(field), uint64(c))
			b, ok := values[value]
			if !ok {
				b = roaring64.New()
				values[value] = b
			}
			b.Add(row)
		})
		m.Read(tx, shard, match, metrics...)
		return nil
	})
	a := &Result{
		Results: make([]map[string]any, 0, len(values)),
	}
	property := o.Name(uint32(field))
	for k, v := range values {
		x := map[string]any{
			property: k,
		}
		for i := range metrics {
			x[metrics[i]] = m.Compute(metrics[i], v)
		}
		a.Results = append(a.Results, x)
	}
	return a, nil
}

func (o *Store) BreakdownExitPages(start, end int64, domain string, filter Filter) (*Result, error) {
	m := NewData()
	defer m.Release()
	values := make(map[string]*roaring64.Bitmap)
	o.Select(start, end, domain, filter, func(tx *Tx, shard uint64, match *roaring64.Bitmap) error {
		tx.ExtractBSI(shard, uint64(alicia.EXIT_PAGE), match, func(row uint64, c int64) {
			value := tx.Find(uint64(alicia.EXIT_PAGE), uint64(c))
			b, ok := values[value]
			if !ok {
				b = roaring64.New()
				values[value] = b
			}
			b.Add(row)
		})
		m.Read(tx, shard, match, visitors, visits, pageviews)
		return nil
	})
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
	sortMap(a.Results, "visitors")
	return a, nil
}

func (o *Store) BreakdownCity(start, end int64, domain string, filter Filter) (*Result, error) {
	values := make(map[uint32]*roaring64.Bitmap)
	m := NewData()
	defer m.Release()
	err := o.Select(start, end, domain, filter, func(tx *Tx, shard uint64, match *roaring64.Bitmap) error {
		tx.ExtractBSI(shard, uint64(alicia.CITY), match, func(row uint64, c int64) {
			code := uint32(c)
			b, ok := values[code]
			if !ok {
				b = roaring64.New()
				values[code] = b
			}
			b.Add(row)
		})
		m.Read(tx, shard, match, visitors)
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
			visitors:       vs,
			"code":         code,
			"name":         city.Name,
			"country_flag": city.Flag,
		})
	}
	sortMap(a.Results, visitors)
	return a, nil
}

func (o *Store) BreakdownVisitorsWithPercentage(start, end int64, domain string, filter Filter, field alicia.Field) (*Result, error) {
	values := make(map[string]*roaring64.Bitmap)
	m := NewData()
	defer m.Release()

	err := o.Select(start, end, domain, filter, func(tx *Tx, shard uint64, match *roaring64.Bitmap) error {
		tx.ExtractBSI(shard, uint64(field), match, func(row uint64, c int64) {
			value := tx.Find(uint64(field), uint64(c))
			b, ok := values[value]
			if !ok {
				b = roaring64.New()
				values[value] = b
			}
			b.Add(row)
		})
		m.Read(tx, shard, match, visitors)
		return nil
	})
	if err != nil {
		return nil, err
	}
	a := &Result{
		Results: make([]map[string]any, 0, len(values)),
	}

	total := m.Compute(visitors, nil)
	property := o.Name(uint32(field))
	for prop, b := range values {
		vs := m.Compute(visitors, b)
		p := float64(0)
		if total != 0 {
			p = (vs / total) * 100.0
		}
		a.Results = append(a.Results, map[string]any{
			property:     prop,
			visitors:     vs,
			"percentage": p,
		})
	}
	sortMap(a.Results, visitors)
	return a, nil

}

func sortMap(ls []map[string]any, key string) {
	slices.SortFunc(ls, func(a, b map[string]any) int {
		return cmp.Compare(b[key].(float64), a[key].(float64))
	})
}
