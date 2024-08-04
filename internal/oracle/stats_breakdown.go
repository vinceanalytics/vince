package oracle

import (
	"cmp"
	"slices"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/gernest/len64/internal/location"
	"github.com/gernest/len64/internal/rbf"
	"github.com/gernest/len64/internal/rbf/cursor"
	"github.com/gernest/roaring"
	"github.com/gernest/rows"
	"go.etcd.io/bbolt"
)

type Breakdown struct {
	Results []map[string]any `json:"results"`
}

func (o *Oracle) Breakdown(start, end int64, domain string, filter Filter, metrics []string, property string) (*Breakdown, error) {
	m := newAggregate()
	values := make(map[string]*roaring64.Bitmap)
	err := o.db.Select(start, end, domain, filter, func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, match *rows.Row) error {
		err := cursor.Tx(rTx, property, func(c *rbf.Cursor) error {
			f := newReadField(tx, []byte(property))
			return extractMutex(c, match, func(row uint64, columns *roaring.Container) {
				value := string(f.read(row))
				b, ok := values[value]
				if !ok {
					b = roaring64.New()
					values[value] = b
				}
				roaring.ContainerCallback(columns, func(u uint16) {
					b.Add(uint64(u))
				})
			})
		})
		if err != nil {
			return err
		}
		return m.read(rTx, shard, match, metrics...)
	})
	if err != nil {
		return nil, err
	}
	a := &Breakdown{
		Results: make([]map[string]any, 0, len(values)),
	}

	ls := make([]string, 0, len(values))
	for k := range values {
		ls = append(ls, k)
	}
	slices.Sort(ls)

	for i := range ls {
		x := map[string]any{
			property: ls[i],
		}
		for _, met := range metrics {
			x[met] = m.Compute(met, values[ls[i]])
		}
		a.Results = append(a.Results, x)
	}
	return a, nil
}

func (o *Oracle) BreakdownCity(start, end int64, domain string, filter Filter) (*Breakdown, error) {
	m := newAggregate()
	values := make(map[uint32]*roaring64.Bitmap)
	err := o.db.Select(start, end, domain, filter, func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, match *rows.Row) error {
		err := cursor.Tx(rTx, "city", func(c *rbf.Cursor) error {
			return extractBSI(c, shard, match, func(column uint64, value int64) error {
				code := uint32(value)
				b, ok := values[code]
				if !ok {
					b = roaring64.New()
					values[code] = b
				}
				b.Add(column)
				return nil
			})
		})
		if err != nil {
			return err
		}
		return m.read(rTx, shard, match, visitors)
	})
	if err != nil {
		return nil, err
	}
	a := &Breakdown{
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

const visitors = "visitors"

func sortMap(ls []map[string]any, key string) {
	slices.SortFunc(ls, func(a, b map[string]any) int {
		return cmp.Compare(b[key].(float64), a[key].(float64))
	})
}

func (o *Oracle) BreakdownVisitorsWithPercentage(start, end int64, domain string, filter Filter, property string) (*Breakdown, error) {
	m := newAggregate()
	values := make(map[string]*roaring64.Bitmap)
	err := o.db.Select(start, end, domain, filter, func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, match *rows.Row) error {
		err := cursor.Tx(rTx, property, func(c *rbf.Cursor) error {
			f := newReadField(tx, []byte(property))
			return extractMutex(c, match, func(row uint64, columns *roaring.Container) {
				value := string(f.read(row))
				b, ok := values[value]
				if !ok {
					b = roaring64.New()
					values[value] = b
				}
				roaring.ContainerCallback(columns, func(u uint16) {
					b.Add(uint64(u))
				})
			})
		})
		if err != nil {
			return err
		}
		return m.read(rTx, shard, match, visitors)
	})
	if err != nil {
		return nil, err
	}
	a := &Breakdown{
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
			"percentage": p,
		})
	}
	sortMap(a.Results, visitors)
	return a, nil
}
