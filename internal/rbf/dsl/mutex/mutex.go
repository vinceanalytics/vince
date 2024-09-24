package mutex

import (
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/cursor"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
)

func Add(m *roaring.Bitmap, id uint64, value uint64) {
	m.Add(value*shardwidth.ShardWidth + (id % shardwidth.ShardWidth))
}

func Extract(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(column, value uint64) error) error {
	return cursor.Rows(c, 0, func(rowID uint64) error {
		row, err := cursor.Row(c, shard, rowID)
		if err != nil {
			return err
		}
		row = row.Intersect(columns)
		return row.RangeColumns(func(u uint64) error {
			return f(u, rowID)
		})
	})
}
