package store

import (
	"time"

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
				tx.NewBitmap(shard, view, models.Field_id).
					ExtractBSI(shard, match, func(_ uint64, value int64) {
						r.Set(uint64(value))
					})
			}
			return nil
		})
	})
	visitors = uint64(r.GetCardinality())
	return
}

func (o *Store) Visitors(domain string) (visitors uint64, err error) {
	err = o.View(func(tx *Tx) error {
		b := roaring.NewBitmap()
		did := []byte(domain)
		for shard := range tx.Shards() {
			match := tx.Domain(shard, 0, did)
			if match.IsEmpty() {
				continue
			}
			tx.NewBitmap(shard, 0, models.Field_id).
				ExtractBSI(shard, match, func(_ uint64, value int64) {
					b.Set(uint64(value))
				})
		}
		visitors = uint64(b.GetCardinality())
		return nil
	})
	return
}
