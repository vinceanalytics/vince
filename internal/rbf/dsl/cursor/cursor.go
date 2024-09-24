package cursor

import (
	"github.com/gernest/rbf"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
)

func Rows(c *rbf.Cursor, start uint64, cb func(row uint64) error, filters ...roaring.BitmapFilter) error {
	startKey := rowToKey(start)
	filter := roaring.NewBitmapRowFilter(cb, filters...)
	return c.ApplyFilter(startKey, filter)
}

func Row(c *rbf.Cursor, shard uint64, rowID uint64) (*rows.Row, error) {
	data, err := c.OffsetRange(
		shard*shardwidth.ShardWidth,
		rowID*shardwidth.ShardWidth,
		(rowID+1)*shardwidth.ShardWidth,
	)
	if err != nil {
		return nil, err
	}
	row := &rows.Row{
		Segments: []rows.RowSegment{
			rows.NewSegment(data, shard, true),
		},
	}
	row.InvalidateCount()
	return row, nil
}

// width of roaring containers is 2^16
const containerWidth = 1 << 16

func rowToKey(rowID uint64) (key uint64) {
	return rowID * (shardwidth.ShardWidth / containerWidth)
}
