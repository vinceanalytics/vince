package bsi

import (
	"math/bits"

	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/cursor"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
)

const (
	// BSI bits used to check existence & sign.
	bsiExistsBit = 0
	bsiSignBit   = 1
	bsiOffsetBit = 2
)

func Add(m *roaring.Bitmap, id uint64, svalue int64) {
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

// Extract finds all values set in exists columns and calls f with the
// found column and value.
func Extract(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(column uint64, value int64) error) error {
	exists, err := cursor.Row(c, shard, bsiExistsBit)
	if err != nil {
		return err
	}
	exists = exists.Intersect(columns)

	data := make(map[uint64]uint64)
	mergeBits(exists, 0, data)
	bitDepth, err := depth(c)
	if err != nil {
		return err
	}
	sign, err := cursor.Row(c, shard, bsiSignBit)
	if err != nil {
		return err
	}
	sign = sign.Intersect(exists)
	mergeBits(sign, 1<<63, data)

	for i := uint64(0); i < bitDepth; i++ {
		bits, err := cursor.Row(c, shard, bsiOffsetBit+uint64(i))
		if err != nil {
			return err
		}
		bits = bits.Intersect(exists)
		mergeBits(bits, 1<<i, data)
	}
	for columnID, val := range data {
		// Convert to two's complement and add base back to value.
		val = uint64((2*(int64(val)>>63) + 1) * int64(val&^(1<<63)))
		err := f(columnID, int64(val))
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeBits(bits *rows.Row, mask uint64, out map[uint64]uint64) {
	for _, v := range bits.Columns() {
		out[v] |= mask
	}
}

func depth(c *rbf.Cursor) (uint64, error) {
	m, err := c.Max()
	return m / rbf.ShardWidth, err
}
