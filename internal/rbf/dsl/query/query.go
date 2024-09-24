package query

import (
	"github.com/gernest/rbf/dsl/tx"
	"github.com/gernest/rows"
)

// Filter  selects rows to read in a shard/view context.
type Filter interface {
	Apply(tx *tx.Tx, columns *rows.Row) (*rows.Row, error)
}

type Noop struct{}

func (Noop) Apply(tx *tx.Tx, columns *rows.Row) (*rows.Row, error) {
	return rows.NewRow(), nil
}

type And []Filter

func (a And) Apply(tx *tx.Tx, columns *rows.Row) (*rows.Row, error) {
	switch len(a) {
	case 0:
		return rows.NewRow(), nil
	case 1:
		return a[0].Apply(tx, columns)
	default:
		r, err := a[0].Apply(tx, columns)
		if err != nil {
			return nil, err
		}

		for _, x := range a[1:] {
			if r.IsEmpty() {
				return r, nil
			}
			n, err := x.Apply(tx, columns)
			if err != nil {
				return nil, err
			}
			r = r.Intersect(n)
		}
		return r, nil
	}
}

type Or []Filter

func (a Or) Apply(tx *tx.Tx, columns *rows.Row) (*rows.Row, error) {
	switch len(a) {
	case 0:
		return rows.NewRow(), nil
	case 1:
		return a[0].Apply(tx, columns)
	default:
		r, err := a[0].Apply(tx, columns)
		if err != nil {
			return nil, err
		}
		for _, x := range a[1:] {
			n, err := x.Apply(tx, columns)
			if err != nil {
				return nil, err
			}
			r = r.Union(n)
		}
		return r, nil
	}
}
