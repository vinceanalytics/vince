package ro2

import (
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (o *Store) CurrentVisitors(domain string) (visitors uint64, err error) {
	end := time.Now().UTC()
	start := end.Add(-5 * time.Minute)
	r := roaring64.New()
	err = o.View(func(tx *Tx) error {
		shard, ok := tx.ID(v1.Field_domain, domain)
		if !ok {
			return nil
		}
		return query.Minute.Range(start, end, func(t time.Time) error {
			view := uint64(t.UnixMilli())
			bs, err := tx.Bitmap(shard, view, v1.Field_domain)
			if err != nil {
				return err
			}
			match := bs.GetExistenceBitmap()
			if match.IsEmpty() {
				return nil
			}
			uniq, err := tx.Transpose(shard, view, v1.Field_id, match)
			if err != nil {
				return err
			}
			r.Or(uniq)
			return nil
		})
	})
	visitors = r.GetCardinality()
	return
}

func (o *Store) Visitors(domain string) (visitors uint64, err error) {
	err = o.View(func(tx *Tx) error {
		shard, ok := tx.ID(v1.Field_domain, domain)
		if !ok {
			return nil
		}

		// use global shard bitmap  to avoid calling comparison. bs is existence
		// bitmap contains all columns  for the shard globally.
		bs, err := tx.Bitmap(shard, 0, v1.Field_domain)
		if err != nil {
			return err
		}
		match := bs.GetExistenceBitmap()
		if match.IsEmpty() {
			return nil
		}
		uniq, err := tx.Unique(0, 0, v1.Field_id, match)
		if err != nil {
			return err
		}
		visitors = uniq
		return nil
	})
	return
}
