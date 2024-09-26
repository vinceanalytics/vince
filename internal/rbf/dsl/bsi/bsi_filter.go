package bsi

import (
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/query"
)

type Match struct {
	field        []byte
	prefix       int
	op           Operation
	valueOrStart int64
	end          int64
}

func Filter(field string, op Operation, valueOrStart int64, end int64) *Match {
	return &Match{
		field:        []byte(field),
		prefix:       len(field),
		op:           op,
		valueOrStart: valueOrStart,
		end:          end,
	}
}

var _ query.Filter = (*Match)(nil)

func (m *Match) Apply(rtx *rbf.Tx, shard uint64, view []byte, columns *rows.Row) (*rows.Row, error) {
	c, err := rtx.Cursor(string(append(m.field[:m.prefix], view...)))
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return Compare(c, shard, m.op, m.valueOrStart, m.end, columns)
}
