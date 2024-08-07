package btx

import (
	"math/bits"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/cursor"
)

func Mutex(m *roaring.Bitmap, id uint64, v uint64) {
	m.DirectAdd(MutexPosition(id, v))
}

func MutexPosition(id uint64, v uint64) uint64 {
	return v*shardwidth.ShardWidth + (id % shardwidth.ShardWidth)
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

func MaxUnsigned(c *rbf.Cursor, filter *rows.Row, shard uint64, bitDepth uint64) (max uint64, count uint64, err error) {
	count = filter.Count()
	for i := int(bitDepth - 1); i >= 0; i-- {
		row, err := cursor.Row(c, shard, uint64(2+i))
		if err != nil {
			return max, count, err
		}
		row = row.Intersect(filter)

		count = row.Count()
		if count > 0 {
			max += (1 << uint(i))
			filter = row
		} else if i == 0 {
			count = filter.Count()
		}
	}
	return max, count, nil
}

func MinUnsigned(c *rbf.Cursor, filter *rows.Row, shard uint64, bitDepth uint64) (min uint64, count uint64, err error) {
	count = filter.Count()
	for i := int(bitDepth - 1); i >= 0; i-- {
		row, err := cursor.Row(c, shard, uint64(2+i))
		if err != nil {
			return min, count, err
		}
		row = filter.Difference(row)
		count = row.Count()
		if count > 0 {
			filter = row
		} else {
			min += (1 << uint(i))
			if i == 0 {
				count = filter.Count()
			}
		}
	}
	return min, count, nil
}

func ExtractBSI(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(column uint64, value int64) error) error {
	exists, err := cursor.Row(c, shard, 0)
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
	sign, err := cursor.Row(c, shard, 1)
	if err != nil {
		return err
	}
	sign = sign.Intersect(exists)
	mergeBits(sign, 1<<63, data)

	for i := uint64(0); i < bitDepth; i++ {
		bits, err := cursor.Row(c, shard, 2+uint64(i))
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
