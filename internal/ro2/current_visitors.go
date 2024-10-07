package ro2

import (
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (o *Store) CurrentVisitors(domain string) (visitors uint64, err error) {
	end := time.Now().UTC()
	start := end.Add(-5 * time.Minute)
	r := roaring.NewBitmap()
	err = o.View(func(tx *Tx) error {
		shard, ok := tx.ID(v1.Field_domain, domain)
		if !ok {
			return nil
		}
		return query.Minute.Range(start, end, func(t time.Time) error {
			view := uint64(t.UnixMilli())
			var match *roaring.Bitmap
			err := tx.Bitmap(shard, view, v1.Field_domain, func(bs *roaring.BSI) {
				match = bs.GetExistenceBitmap().Clone()
			})
			if err != nil {
				return err
			}
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
	visitors = uint64(r.GetCardinality())
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
		var match *roaring.Bitmap
		err := tx.Bitmap(shard, 0, v1.Field_domain, func(bs *roaring.BSI) {
			match = bs.GetExistenceBitmap().Clone()
		})
		if err != nil {
			return err
		}
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
