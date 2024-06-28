package db

import (
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rows"
)

type realtimeQuery struct {
	roaring64.Bitmap
	fmt ViewFmt
}

func (r *realtimeQuery) View(_ time.Time) View {
	return r
}

func (r *realtimeQuery) Apply(tx *Tx, columns *rows.Row) error {
	view := r.fmt.Format(tx.View, "uid")
	return transpose(tx.Tx, &r.Bitmap, view, tx.Shard, columns)
}

func (r *realtimeQuery) Visitors() uint64 {
	return r.GetCardinality()
}
