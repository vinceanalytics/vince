package ro2

import (
	"errors"
	"io"
	"time"

	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func (o *Store) CurrentVisitors(domain string) (visitors uint64, err error) {
	end := time.Now().UTC()
	start := end.Add(5 * time.Minute)
	r := roaring64.New()
	err = o.Select(
		start.UnixMilli(), end.UnixMilli(), domain, nil,
		func(tx *Tx, shard uint64, match *roaring64.Bitmap) error {
			tx.ExtractBSI(shard, uint64(alicia.ID), match, func(row uint64, c int64) {
				r.Add(uint64(c))
			})
			// Only process a single shard. It doesn't matter if this query span more
			// than a single shard.
			return io.EOF
		},
	)
	if errors.Is(err, io.EOF) {
		err = nil
	}
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
