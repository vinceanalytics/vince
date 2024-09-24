package boolean

import (
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/cursor"
)

const (
	// Row ids used for boolean fields.
	falseRowID = uint64(0)
	trueRowID  = uint64(1)

	falseRowOffset = 0 * shardwidth.ShardWidth // fragment row 0
	trueRowOffset  = 1 * shardwidth.ShardWidth // fragment row 1
)

func Add(id uint64, value bool) uint64 {
	fragmentColumn := id % shardwidth.ShardWidth
	if value {
		return trueRowOffset + fragmentColumn
	}
	return falseRowOffset + fragmentColumn
}

func Extract(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(column uint64, value bool) error) error {
	trueRow, err := cursor.Row(c, shard, trueRowID)
	if err != nil {
		return err
	}
	falseRow, err := cursor.Row(c, shard, falseRowID)
	if err != nil {
		return err
	}

	return columns.RangeColumns(func(u uint64) error {
		switch {
		case trueRow.Includes(u):
			return f(u, true)
		case falseRow.Includes(u):
			return f(u, false)
		default:
			return nil
		}
	})
}
