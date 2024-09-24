package boolean

import (
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/cursor"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/query"
)

type Match struct {
	field string
	value bool
}

func Filter(field string, value bool) *Match {
	return &Match{field: field, value: value}
}

var _ query.Filter = (*Match)(nil)

func (m *Match) Apply(rtx *rbf.Tx, shard uint64, columns *rows.Row) (*rows.Row, error) {
	c, err := rtx.Cursor(m.field)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var r *rows.Row
	if m.value {
		r, err = cursor.Row(c, shard, trueRowID)
	} else {
		r, err = cursor.Row(c, shard, falseRowID)
	}
	if err != nil {
		return nil, err
	}
	if columns != nil {
		r = r.Intersect(columns)
	}
	return r, nil
}
