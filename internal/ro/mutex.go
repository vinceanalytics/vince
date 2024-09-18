package ro

import "github.com/vinceanalytics/vince/internal/roaring/roaring64"

func Mutex(m *roaring64.Bitmap, id uint64, value uint64) {
	m.Add(value*ShardWidth + (id % ShardWidth))
}
