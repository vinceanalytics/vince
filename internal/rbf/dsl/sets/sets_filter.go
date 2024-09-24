package sets

import (
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/cursor"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/query"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/tx"
)

type Match struct {
	field string
	value uint64
}

func Filter(field string, value uint64) *Match {
	return &Match{
		field: field,
		value: value,
	}
}

var _ query.Filter = (*Match)(nil)

func (m *Match) Apply(tx *tx.Tx, columns *rows.Row) (*rows.Row, error) {
	c, err := tx.Get(m.field)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	r, err := cursor.Row(c, tx.Shard(), m.value)
	if err != nil {
		return nil, err
	}
	if columns != nil {
		r = r.Intersect(columns)
	}
	return r, nil
}
