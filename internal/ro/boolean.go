package ro

import "github.com/vinceanalytics/vince/internal/roaring/roaring64"

func True(m *roaring64.Bitmap, id uint64) {
	m.Add(id % ShardWidth)
}

func False(m *roaring64.Bitmap, id uint64) {
	m.Add(ShardWidth + (id % ShardWidth))
}
