package ro2

import (
	"time"

	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/cursor"
	"github.com/vinceanalytics/vince/internal/rbf/quantum"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (o *Store) Stats(domain string, start, end time.Time, interval query.Interval, filters query.Filters, metrics []string) (*Stats, error) {
	m := new(Stats)
	fields := MetricsToProject(metrics)

	err := o.View(func(tx *Tx) error {
		domainId, ok := tx.ID(uint64(alicia.DOMAIN), domain)
		if !ok {
			return nil
		}
		shards := o.Shards(tx)
		match := tx.compile(filters)
		for _, shard := range shards {
			err := o.shards.View(shard, func(rtx *rbf.Tx) error {
				f := quantum.NewField()
				defer f.Release()
				var cb func(name string, start time.Time, end time.Time, fn func([]byte) error) error
				switch interval {
				case query.Minute:
					cb = f.Minute
				case query.Hour:
					cb = f.Hour
				case query.Week:
					cb = f.Week
				case query.Month:
					cb = f.Month
				default:
					cb = f.Day
				}

				return cb(domainField, start, end, func(b []byte) error {
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
	return m, nil
}
