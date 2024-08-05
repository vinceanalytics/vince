package btx

import (
	"math/bits"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
)

func Mutex(m *roaring.Bitmap, id uint64, v uint64) {
	m.DirectAdd(v*shardwidth.ShardWidth + (id % shardwidth.ShardWidth))
}

func BSI(m *roaring.Bitmap, id uint64, svalue int64) {
	fragmentColumn := id % shardwidth.ShardWidth
	m.DirectAdd(fragmentColumn)
	negative := svalue < 0
	var value uint64
	if negative {
		m.Add(shardwidth.ShardWidth + fragmentColumn) // set sign bit
		value = uint64(svalue * -1)
	} else {
		value = uint64(svalue)
	}
	lz := bits.LeadingZeros64(value)
	row := uint64(2)
	for mask := uint64(0x1); mask <= 1<<(64-lz) && mask != 0; mask = mask << 1 {
		if value&mask > 0 {
			m.DirectAdd(row*shardwidth.ShardWidth + fragmentColumn)
		}
		row++
	}
}
