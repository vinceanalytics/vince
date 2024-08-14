package ro

import (
	"math/bits"

	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

func BSI(m *roaring64.Bitmap, id uint64, svalue int64) {
	fragmentColumn := id % ShardWidth
	m.Add(fragmentColumn)
	negative := svalue < 0
	var value uint64
	if negative {
		m.Add(ShardWidth + fragmentColumn) // set sign bit
		value = uint64(svalue * -1)
	} else {
		value = uint64(svalue)
	}
	lz := bits.LeadingZeros64(value)
	row := uint64(2)
	for mask := uint64(0x1); mask <= 1<<(64-lz) && mask != 0; mask = mask << 1 {
		if value&mask > 0 {
			m.Add(row*ShardWidth + fragmentColumn)
		}
		row++
	}
}
