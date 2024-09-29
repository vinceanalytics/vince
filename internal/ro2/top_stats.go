package ro2

import (
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	"github.com/vinceanalytics/vince/internal/fieldset"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (o *Store) Stats(domain string, start, end time.Time, interval query.Interval, filters query.Filters, metrics []string) (*Stats, error) {
	m := new(Stats)
	fields := fieldset.From(metrics...)
	err := o.View(func(tx *Tx) error {
		return tx.Select(domain, start, end, interval, filters, func(shard, view uint64, columns *roaring64.Bitmap) error {
			return m.Read(tx, shard, view, columns, fields)
		})
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}
