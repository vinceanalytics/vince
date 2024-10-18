package ro2

import (
	"time"

	"github.com/vinceanalytics/vince/internal/bsi"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (o *Store) CurrentVisitors(domain string) (visitors uint64, err error) {
	end := time.Now().UTC()
	start := end.Add(-5 * time.Minute)
	r := roaring.NewBitmap()
	err = o.View(func(tx *Tx) error {
		return query.Minute.Range(start, end, func(t time.Time) error {
			view := uint64(t.UnixMilli())
			did := []byte(domain)
			for shard := range tx.Shards() {
				match := tx.Domain(shard, view, did)
				if match.IsEmpty() {
					continue
				}
				uniq := tx.Transpose(shard, view, models.Field_id, match)
				r.Or(uniq)
			}
			return nil
		})
	})
	visitors = uint64(r.GetCardinality())
	return
}

func (o *Store) Visitors(domain string) (visitors uint64, err error) {
	err = o.View(func(tx *Tx) error {
		b := bsi.NewBitmap()
		did := []byte(domain)
		for shard := range tx.Shards() {
			match := tx.Domain(shard, 0, did)
			if match.IsEmpty() {
				continue
			}
			uniq := tx.Transpose(shard, 0, models.Field_id, match)
			b.Or(uniq)
		}
		visitors = uint64(b.GetCardinality())
		return nil
	})
	return
}
