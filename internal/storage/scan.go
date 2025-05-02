package storage

import (
	"iter"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/storage/bitmaps"
	"github.com/vinceanalytics/vince/internal/storage/cursor"
	"github.com/vinceanalytics/vince/internal/util/shard"
)

type Filter interface {
	Apply(shard *shard.Shard) *ro2.Bitmap
}

func (b *Handle) Scan(
	domain []byte,
	res encoding.Resolution,
	start, end time.Time,
	filter Filter,
	valueSet models.BitSet,
) iter.Seq[*shard.Shard] {
	domainID := b.mapping.Get(v1.Field_domain, domain)
	if domainID == 0 {
		return func(yield func(*shard.Shard) bool) {}
	}

	return func(yield func(*shard.Shard) bool) {
		cu := cursor.New(nil, 0)
		sx := new(shard.Shard)
		for sha := range b.Shards() {
			domains := bitmaps.Row(cu, sha, domainID)
			if !domains.Any() {
				continue
			}
			resolved := cu.Resolve(res, sha, start, end)

			for i := range resolved.Columns {
				ra := resolved.Columns[i].Intersect(domains)
				if !ra.Any() {
					continue
				}

				ra = filter.Apply(sx)
				if !ra.Any() {
					continue
				}
				*sx = shard.Shard{
					View:    resolved.Timestamp[i],
					Shard:   sha,
					Mapping: b.mapping,
				}
				if !yield(sx) {
					return
				}
			}
		}
	}
}
