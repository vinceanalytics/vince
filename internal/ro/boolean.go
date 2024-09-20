package ro

import "github.com/vinceanalytics/vince/internal/roaring/roaring64"

func Bool(m *roaring64.Bitmap, id uint64) {
	m.Add(id % ShardWidth)
}
