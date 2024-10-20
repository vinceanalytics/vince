package store

import (
	"time"

	"github.com/vinceanalytics/vince/internal/fieldset"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (o *Store) Timeseries(domain string, params *query.Query, metrics []string) (map[string]*Stats, error) {
	values := make(map[string]*Stats)
	fields := fieldset.From(metrics...)

	err := o.View(func(tx *Tx) error {
		format := params.Interval().Format()
		return tx.Select(domain, params.Start(), params.End(), params.Interval(), params.Filter(), func(shard, view uint64, columns *roaring.Bitmap) error {
			timestamp := time.UnixMilli(int64(view)).UTC().Format(format)
			m, ok := values[timestamp]
			if !ok {
				m = NewStats(fields)
				values[timestamp] = m
			}
			return m.Read(tx, shard, view, columns, fields)
		})
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}
