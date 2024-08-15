package ro2

import (
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

const (
	visitors = "visitors"
)

type Result struct {
	Results []map[string]any `json:"results"`
}

func (o *Proto[T]) BreakdownVisitorsWithPercentage(start, end int64, domain string, filter Filter, field uint32) (*Result, error) {
	values := make(map[string]*roaring64.Bitmap)
	var m Data

	err := o.Select(start, end, domain, filter, func(tx *Tx, shard uint64, match *roaring64.Bitmap) error {
		tx.ExtractMutex(shard, uint64(field), match, func(row uint64, c *roaring.Container) {
			value := tx.Find(uint32(row))
			b, ok := values[value]
			if !ok {
				b = roaring64.New()
				values[value] = b
			}
			c.Each(func(u uint16) bool {
				b.Add(uint64(u))
				return true
			})
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
	property := o.fields[field]
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
	return a, nil

}
