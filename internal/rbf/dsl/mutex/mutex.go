package mutex

import (
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
)

func Add(id uint64, value uint64) uint64 {
	return value*shardwidth.ShardWidth + (id % shardwidth.ShardWidth)
}

func Extract(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(column, value uint64) error) error {
	return Distinct(c, columns, func(row uint64, columns *roaring.Container) error {
		var err error
		roaring.ContainerCallback(columns, func(u uint16) {
			if err != nil {
				return
			}
			err = f(uint64(u), row)
		})
		return err
	})
}
