package db

import (
	"errors"

	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/cursor"
	"github.com/gernest/rows"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

func (tx *view) filters(list []*v1.Filter, f *rows.Row) (*rows.Row, error) {
	if len(list) == 0 {
		return f, nil
	}
	var err error
	fs := make([]*rows.Row, len(list))
	for i := range list {
		fs[i], err = tx.filter(list[i])
		if err != nil {
			return nil, err
		}
	}
	if len(fs) == 1 {
		return fs[0].Intersect(f), nil
	}
	return fs[0].Union(fs[1:]...).Intersect(f), nil
}

func (tx *view) filter(filter *v1.Filter) (*rows.Row, error) {
	switch filter.Op {
	case v1.Filter_equal:
		return tx.eq(filter.Property.String(), filter.Value)
	case v1.Filter_not_equal:
		return tx.neq(filter.Property.String(), filter.Value)
	case v1.Filter_re_equal:
		return tx.re(filter.Property.String(), filter.Value)
	case v1.Filter_re_not_equal:
		return tx.nre(filter.Property.String(), filter.Value)
	default:
		return rows.NewRow(), nil
	}
}

func (tx *view) eq(field, value string) (*rows.Row, error) {
	return eq(tx, tx.shard, field, value)
}

func (tx *view) neq(field, value string) (*rows.Row, error) {
	ex, err := tx.ids()
	if err != nil {
		return nil, err
	}
	r, err := tx.eq(field, value)
	if err != nil {
		return nil, err
	}
	return ex.Difference(r), nil
}

func (tx *view) ids() (*rows.Row, error) {
	c, err := tx.get("_id")
	if err != nil {
		return nil, err
	}
	return cursor.Row(c, tx.shard, 0)
}

func (tx *view) nre(field, value string) (*rows.Row, error) {
	ex, err := tx.ids()
	if err != nil {
		return nil, err
	}
	r, err := tx.re(field, value)
	if err != nil {
		return nil, err
	}
	return ex.Difference(r), nil
}

func (tx *view) re(key, value string) (*rows.Row, error) {
	c, err := tx.get(key)
	if err != nil {
		if errors.Is(err, rbf.ErrBitmapNotFound) {
			return rows.NewRow(), nil
		}
		return nil, err
	}
	match := tx.findRe(key, value)
	if len(match) == 0 {
		return rows.NewRow(), nil
	}
	matches := make([]*rows.Row, len(match))
	for i := range match {
		matches[i], err = cursor.Row(c, tx.shard, match[i])
		if err != nil {
			return nil, err
		}
	}
	return matches[0].Union(matches[1:]...), nil
}
