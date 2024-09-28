package ro2

import (
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (o *Store) CurrentVisitors(domain string) (visitors uint64, err error) {
	end := time.Now().UTC().Truncate(time.Minute)
	start := end.Add(-5 * time.Minute)
	m := new(Stats)
	err = o.View(func(tx *Tx) error {
		return tx.Select(domain, start, end, query.Minute, query.Filters{}, func(shard, view uint64, columns *roaring64.Bitmap) error {
			return m.ReadFields(tx, shard, view, columns, v1.Field_id)
		})
	})
	visitors = uint64(m.Visitors)
	return
}

func (o *Store) Visitors(domain string) (visitors uint64, err error) {
	err = o.View(func(tx *Tx) error {
		shard, ok := tx.ID(v1.Field_domain, domain)
		if !ok {
			return nil
		}
		bs, err := tx.Bitmap(0, 0, v1.Field_domain)
		if err != nil {
			return err
		}
		match := bs.CompareValue(0, roaring64.EQ, int64(shard), 0, bs.GetExistenceBitmap())
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
