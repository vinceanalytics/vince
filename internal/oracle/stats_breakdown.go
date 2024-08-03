package oracle

import (
	"github.com/RoaringBitmap/roaring/v2/roaring64"
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
	for k, v := range values {
		x := map[string]any{
			property: k,
		}
		for _, met := range metrics {
			x[met] = m.Compute(met, v)
		}
		a.Results = append(a.Results, x)
	}
	return a, nil
}
