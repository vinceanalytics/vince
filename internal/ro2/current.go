package ro2

import (
	"hash/crc32"

	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func (o *Proto[T]) Today(domain string) (visitors uint64, err error) {
	date := uint64(ro.Today().UnixMilli())
	shards := o.shards()
	hash := crc32.NewIEEE()
	hash.Write([]byte(domain))
	domainID := uint64(hash.Sum32())
	uniq := roaring64.NewBitmap()
	err = o.View(func(tx *Tx) error {
		for i := range shards {
			shard := shards[i]
			b := tx.Row(shard, domainField, domainID)
			if b.IsEmpty() {
				continue
			}
			tx.ExtractBSI(date, shard, idField, 64, b, func(row uint64, c int64) {
				uniq.Add(uint64(c))
			})
		}
		return nil
	})
	visitors = uniq.GetCardinality()
	return
}
