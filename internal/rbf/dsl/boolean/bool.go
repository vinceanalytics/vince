package boolean

import (
	"errors"

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

func Extract(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(column uint64, value bool)) error {
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
			f(u, true)
		case falseRow.Includes(u):
			f(u, false)
		}
		return nil
	})
}

func Bounce(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(column uint64, value int64)) error {
	trueRow, err := cursor.Row(c, shard, trueRowID)
	if err != nil {
		return err
	}
	falseRow, err := cursor.Row(c, shard, falseRowID)
	if err != nil {
		return err
	}
	if columns != nil {
		falseRow = falseRow.Intersect(columns)
		trueRow = trueRow.Intersect(columns)
	}
	return errors.Join(
		trueRow.RangeColumns(func(u uint64) error {
			f(u, 1)
			return nil
		}),
		falseRow.RangeColumns(func(u uint64) error {
			f(u, -1)
			return nil
		}),
	)

}

func BounceCount(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(value int64) error) error {
	trueRow, err := cursor.Row(c, shard, trueRowID)
	if err != nil {
		return err
	}
	falseRow, err := cursor.Row(c, shard, falseRowID)
	if err != nil {
		return err
	}
	if columns != nil {
		falseRow = falseRow.Intersect(columns)
		trueRow = trueRow.Intersect(columns)
	}
	return f(int64(trueRow.Count()) + -(int64(falseRow.Count())))
}

func True(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(column uint64, value int64)) error {
	trueRow, err := cursor.Row(c, shard, trueRowID)
	if err != nil {
		return err
	}
	if columns != nil {
		trueRow = trueRow.Intersect(columns)
	}
	return trueRow.RangeColumns(func(u uint64) error {
		f(u, 1)
		return nil
	})
}

func Count(c *rbf.Cursor, shard uint64, columns *rows.Row, f func(value int64) error) error {
	trueRow, err := cursor.Row(c, shard, trueRowID)
	if err != nil {
		return err
	}
	if columns != nil {
		trueRow = trueRow.Intersect(columns)
	}
	return f(int64(trueRow.Count()))
}
