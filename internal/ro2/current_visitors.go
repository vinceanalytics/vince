package ro2

import (
	"time"

	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func (o *Store) CurrentVisitors(domain string) (visitors uint64, err error) {
	end := time.Now().UTC().Truncate(time.Minute)
	// we want the currenttime to be included in the search
	end = end.Add(time.Minute)
	start := end.Add(-5 * time.Minute)

	shard := o.seq.Load() / ro.ShardWidth

	dom := NewEq(uint64(alicia.DOMAIN), domain)
	r := roaring64.New()

	err = o.View(func(tx *Tx) error {
		b := dom.match(tx, shard)
		if b.IsEmpty() {
			return nil
		}
		ts := tx.Cmp(uint64(alicia.MINUTE), shard, roaring64.RANGE,
			start.UnixMilli(), end.UnixMilli(),
		)
		if ts.IsEmpty() {
			return nil
		}
		b.And(ts)
		if b.IsEmpty() {
			return nil
		}
		tx.ExtractBSI(shard, uint64(alicia.ID), b, func(row uint64, c int64) {
			r.Add(uint64(c))
		})
		return nil
	})
	visitors = r.GetCardinality()
	return
}

func (o *Store) Visitors(domain string) (visitors uint64, err error) {
	r := roaring64.New()
	dom := NewEq(uint64(alicia.DOMAIN), domain)
	shards := o.shards()
	err = o.View(func(tx *Tx) error {
		for i := range shards {
			shard := shards[i]
			b := dom.match(tx, shard)
			if b.IsEmpty() {
				continue
			}
			tx.ExtractBSI(shard, uint64(alicia.ID), b, func(row uint64, c int64) {
				// we don't care about row we just neeed to find unique users
				r.Add(uint64(c))
			})
		}
		return nil
	})
	visitors = r.GetCardinality()
	return
}
