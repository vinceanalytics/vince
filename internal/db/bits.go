package db

import (
	"math/bits"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
)

const (
	falseRowOffset = 0 * shardwidth.ShardWidth // fragment row 0
	trueRowOffset  = 1 * shardwidth.ShardWidth // fragment row 1
)

func mutex(m *roaring.Bitmap, id uint64, value uint64) {
	m.Add(value*shardwidth.ShardWidth + (id % shardwidth.ShardWidth))
}

func bsi(m *roaring.Bitmap, id, value uint64) {
	fragmentColumn := id % shardwidth.ShardWidth
	m.DirectAdd(fragmentColumn)
	lz := bits.LeadingZeros64(value)
	row := uint64(2)
	for mask := uint64(0x1); mask <= 1<<(64-lz) && mask != 0; mask = mask << 1 {
		if value&mask > 0 {
			m.DirectAdd(row*shardwidth.ShardWidth + fragmentColumn)
		}
		row++
	}
}

func boolean(m *roaring.Bitmap, id uint64, value bool) {
	fragmentColumn := id % shardwidth.ShardWidth
	if value {
		m.DirectAdd(trueRowOffset + fragmentColumn)
	} else {
		m.DirectAdd(falseRowOffset + fragmentColumn)
	}
}
