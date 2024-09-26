package query

import (
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
)

// Filter  selects rows to read in a shard/view context.
type Filter interface {
	Apply(rtx *rbf.Tx, shard uint64, view []byte, columns *rows.Row) (*rows.Row, error)
}

type FilterFn func(rtx *rbf.Tx, shard uint64, view []byte, columns *rows.Row) (*rows.Row, error)

func (f FilterFn) Apply(rtx *rbf.Tx, shard uint64, view []byte, columns *rows.Row) (*rows.Row, error) {
	return f(rtx, shard, view, columns)
}

type Noop struct{}

func (Noop) Apply(rtx *rbf.Tx, shard uint64, view []byte, columns *rows.Row) (*rows.Row, error) {
	return rows.NewRow(), nil
}

type And []Filter

func (a And) Apply(rtx *rbf.Tx, shard uint64, view []byte, columns *rows.Row) (*rows.Row, error) {
	switch len(a) {
	case 0:
		return columns, nil
	case 1:
		return a[0].Apply(rtx, shard, view, columns)
	default:
		r, err := a[0].Apply(rtx, shard, view, columns)
		if err != nil {
			return nil, err
		}

		for _, x := range a[1:] {
			if r.IsEmpty() {
				return r, nil
			}
			n, err := x.Apply(rtx, shard, view, columns)
			if err != nil {
				return nil, err
			}
			r = r.Intersect(n)
		}
		return r, nil
	}
}

type Or []Filter

func (a Or) Apply(rtx *rbf.Tx, shard uint64, view []byte, columns *rows.Row) (*rows.Row, error) {
	switch len(a) {
	case 0:
		return columns, nil
	case 1:
		return a[0].Apply(rtx, shard, view, columns)
	default:
		r, err := a[0].Apply(rtx, shard, view, columns)
		if err != nil {
			return nil, err
		}
		for _, x := range a[1:] {
			n, err := x.Apply(rtx, shard, view, columns)
			if err != nil {
				return nil, err
			}
			r = r.Union(n)
		}
		return r, nil
	}
}
