package oracle

import (
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/cursor"
)

func extractBSI(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(column uint64, value int64) error) error {
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

func MinMax(c *rbf.Cursor, shard uint64) (int64, int64, error) {
	consider, err := cursor.Row(c, shard, bsiExistsBit)
	if err != nil {
		return 0, 0, err
	}
	min, _, err := minUnsigned(c, consider, shard, 64)
	if err != nil {
		return 0, 0, err
	}
	max, _, err := maxUnsigned(c, consider, shard, 64)
	if err != nil {
		return 0, 0, err
	}
	return int64(min), int64(max), nil

}

func maxUnsigned(c *rbf.Cursor, filter *rows.Row, shard uint64, bitDepth uint64) (max uint64, count uint64, err error) {
	count = filter.Count()
	for i := int(bitDepth - 1); i >= 0; i-- {
		row, err := cursor.Row(c, shard, uint64(bsiOffsetBit+i))
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

func minUnsigned(c *rbf.Cursor, filter *rows.Row, shard uint64, bitDepth uint64) (min uint64, count uint64, err error) {
	count = filter.Count()
	for i := int(bitDepth - 1); i >= 0; i-- {
		row, err := cursor.Row(c, shard, uint64(bsiOffsetBit+i))
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