package ro2

import (
	"errors"
	"io"
	"time"

	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func (o *Proto[T]) CurrentVisitors(domain string) (visitors uint64, err error) {
	end := time.Now().UTC()
	start := end.Add(5 * time.Minute)
	r := roaring64.New()
	err = o.Select(
		start.UnixMilli(), end.UnixMilli(), domain, nil,
		func(tx *Tx, shard uint64, match *roaring64.Bitmap) error {
			tx.ExtractBSI(shard, idField, 64, match, func(row uint64, c int64) {
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
