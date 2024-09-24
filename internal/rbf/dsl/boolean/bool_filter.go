package boolean

import (
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/cursor"
	"github.com/gernest/rbf/dsl/query"
	"github.com/gernest/rbf/dsl/tx"
	"github.com/gernest/rows"
)

type Match struct {
	field string
	value bool
}

func Filter(field string, value bool) *Match {
	return &Match{field: field, value: value}
}

var _ query.Filter = (*Match)(nil)

func (m *Match) Apply(tx *tx.Tx, columns *rows.Row) (*rows.Row, error) {
	c, err := tx.Get(m.field)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var r *rows.Row
	if m.value {
		r, err = cursor.Row(c, tx.Shard(), trueRowID)
	} else {
		r, err = cursor.Row(c, tx.Shard(), falseRowID)
	}
	if err != nil {
		return nil, err
	}
	if columns != nil {
		r = r.Intersect(columns)
	}
	return r, nil
}

func Count(txn *tx.Tx, field string, isTrue bool, columns *rows.Row) (count int64, err error) {
	err = txn.Cursor(field, func(c *rbf.Cursor, tx *tx.Tx) error {
		var r *rows.Row
		var err error
		if isTrue {
			r, err = cursor.Row(c, tx.Shard(), trueRowID)

		} else {
			r, err = cursor.Row(c, tx.Shard(), falseRowID)
		}
		if err != nil {
			return err
		}
		if columns != nil {
			r = r.Intersect(columns)
		}
		count = int64(r.Count())
		return nil
	})
	return
}
