package bsi

import (
	"github.com/gernest/rbf/dsl/query"
	"github.com/gernest/rbf/dsl/tx"
	"github.com/gernest/rows"
)

type Match struct {
	field        string
	op           Operation
	valueOrStart int64
	end          int64
}

func Filter(field string, op Operation, valueOrStart int64, end int64) *Match {
	return &Match{
		field:        field,
		op:           op,
		valueOrStart: valueOrStart,
		end:          end,
	}
}

var _ query.Filter = (*Match)(nil)

func (m *Match) Apply(tx *tx.Tx, columns *rows.Row) (*rows.Row, error) {
	c, err := tx.Get(m.field)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return Compare(c, tx.Shard(), m.op, m.valueOrStart, m.end, columns)
}
