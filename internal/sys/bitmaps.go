package sys

import (
	"github.com/gernest/roaring"
	"github.com/vinceanalytics/vince/internal/btx"
)

type bitmaps struct {
	m map[string]*roaring.Bitmap
}

func newBitmaps() *bitmaps {
	return &bitmaps{m: make(map[string]*roaring.Bitmap)}
}

func (b *bitmaps) Counter(name string, id uint64, g *Counter) {
	btx.BSI(b.get(name), id, g.v.Load())
}

func (b *bitmaps) Gauge(name string, id uint64, g *Gauge) {
	if g.f == nil {
		return
	}
	btx.BSI(b.get(name+"_sum"), id, fromFloat(g.get()))
}

func (b *bitmaps) Histogram(name string, id uint64, h *Histogram) {
	bucketSet := b.get(name + "_buckets_set")
	bucketCount := b.get(name + "_buckets_count")
	total, sum := h.Marshal(func(bucket int, count uint64) {
		btx.Mutex(bucketSet, id, uint64(bucket))
		btx.Mutex(bucketCount, id, count)
	})
	btx.Mutex(b.get(name+"_total"), id, total)
	btx.BSI(b.get(name+"_sum"), id, fromFloat(sum))
}

func (b *bitmaps) get(name string) *roaring.Bitmap {
	m, ok := b.m[name]
	if !ok {
		m = roaring.NewBitmap()
		b.m[name] = m
	}
	return m
}
