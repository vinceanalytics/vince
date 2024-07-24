package len64

import (
	"fmt"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
)

func (db *Store[T]) CurrentVisitor(domain string, duration time.Duration) (uint64, error) {
	if duration == 0 {
		duration = 5 * time.Minute
	}
	end := time.Now().UTC()
	start := end.
		Add(-duration).
		Truncate(time.Second)
	var count uint64
	err := db.View(start, end, domain, func(db *View, shard uint64, match *roaring64.Bitmap) error {
		id, err := db.Get("uid")
		if err != nil {
			return fmt.Errorf("reading id bsi%w", err)
		}
		uniq := id.TransposeWithCounts(
			parallel(), match, match)
		count += uniq.GetCardinality()
		return nil
	})
	if err != nil {
		return 0, err
	}
	return 0, nil
}
