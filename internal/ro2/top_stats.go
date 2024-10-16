package ro2

import (
	"time"

	"github.com/vinceanalytics/vince/internal/fieldset"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (o *Store) Aggregates(domain string, start, end time.Time, interval query.Interval, filters query.Filters, metrics []string) (*Stats, error) {
	fields := fieldset.From(metrics...)
	m := NewStats(fields)
	err := o.View(func(tx *Tx) error {
		return tx.Select(domain, start, end, interval, filters, func(shard, view uint64, columns *roaring.Bitmap) error {
			return m.Read(tx, shard, view, columns, fields)
		})
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}
