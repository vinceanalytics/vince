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
