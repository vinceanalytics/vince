package mutex

import (
	"errors"
	"io"

	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/cursor"
	"github.com/gernest/rbf/dsl/query"
	"github.com/gernest/rbf/dsl/tx"
	"github.com/gernest/rows"
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

type OP uint

const (
	EQ OP = iota
	NEQ
	RE
	NRE
)

type MatchString struct {
	Field string
	Op    OP
	Value string
}

func (m *MatchString) Apply(txn *tx.Tx, columns *rows.Row) (r *rows.Row, err error) {
	err = txn.Cursor(m.Field, func(c *rbf.Cursor, tx *tx.Tx) error {
		switch m.Op {
		case EQ:
			r, err = eq(c, m.Field, []byte(m.Value), tx, columns)
		case NEQ:
			r, err = neq(c, m.Field, []byte(m.Value), tx, columns)
		case RE:
			r, err = m.re(c, tx, columns)
		case NRE:
			r, err = m.nre(c, tx, columns)
		default:
			r = rows.NewRow()
		}
		return err
	})
	return
}

func neq(c *rbf.Cursor, field string, key []byte, tx *tx.Tx, columns *rows.Row) (r *rows.Row, err error) {
	exists, err := existence(tx)
	if err != nil {
		return nil, err
	}
	r, err = eq(c, field, key, tx, columns)
	if err != nil {
		return nil, err
	}
	return exists.Difference(r), nil
}

func neqBlob(c *rbf.Cursor, field string, key []byte, tx *tx.Tx, columns *rows.Row) (r *rows.Row, err error) {
	exists, err := existence(tx)
	if err != nil {
		return nil, err
	}
	r, err = eqBlob(c, field, key, tx, columns)
	if err != nil {
		return nil, err
	}
	return exists.Difference(r), nil
}

func existence(txn *tx.Tx) (r *rows.Row, err error) {
	err = txn.Cursor("_id", func(c *rbf.Cursor, tx *tx.Tx) error {
		r, err = cursor.Row(c, tx.Shard(), 0)
		return err
	})
	return
}

func eq(c *rbf.Cursor, field string, key []byte, tx *tx.Tx, columns *rows.Row) (r *rows.Row, err error) {
	id, ok := tx.Find(field, key)
	if !ok {
		return rows.NewRow(), nil
	}
	return eqID(c, tx, id, columns)
}

func eqBlob(c *rbf.Cursor, field string, key []byte, tx *tx.Tx, columns *rows.Row) (r *rows.Row, err error) {
	id, ok := tx.Find(field, key)
	if !ok {
		return rows.NewRow(), nil
	}
	return eqID(c, tx, id, columns)
}

func eqID(c *rbf.Cursor, tx *tx.Tx, id uint64, columns *rows.Row) (r *rows.Row, err error) {
	r, err = cursor.Row(c, tx.Shard(), id)
	if err != nil {
		return
	}
	if columns != nil {
		r = r.Intersect(columns)
	}
	return
}

func (m *MatchString) nre(c *rbf.Cursor, tx *tx.Tx, columns *rows.Row) (r *rows.Row, err error) {
	exists, err := existence(tx)
	if err != nil {
		return nil, err
	}
	r, err = m.re(c, tx, columns)
	if err != nil {
		return nil, err
	}
	return exists.Difference(r), nil
}

func (m *MatchString) re(c *rbf.Cursor, tx *tx.Tx, columns *rows.Row) (r *rows.Row, err error) {
	err = tx.SearchRe(m.Field, m.Value, nil, nil, func(key []byte, value uint64) error {
		n, err := eqID(c, tx, value, columns)
		if err != nil {
			return err
		}
		if r != nil {
			r = n
		} else {
			r = r.Intersect(n)
		}
		if r.IsEmpty() {
			return io.EOF
		}
		return nil
	})
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return
}

type Blob struct {
	Field string
	Op    OP
	Value []byte
}

var _ query.Filter = (*Blob)(nil)

func (b *Blob) Apply(txn *tx.Tx, columns *rows.Row) (r *rows.Row, err error) {
	err = txn.Cursor(b.Field, func(c *rbf.Cursor, tx *tx.Tx) error {
		switch b.Op {
		case EQ:
			r, err = eqBlob(c, b.Field, b.Value, txn, columns)
		case NEQ:
			r, err = neqBlob(c, b.Field, b.Value, txn, columns)
		default:
			r = rows.NewRow()
		}
		return err
	})
	return
}
