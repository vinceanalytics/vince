package oracle

import (
	"regexp"
	"slices"

	"github.com/gernest/roaring"
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/cursor"
	"go.etcd.io/bbolt"
)

type Filter interface {
	Apply(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (*rows.Row, error)
}

type filterFunc func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (*rows.Row, error)

func (fn filterFunc) Apply(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (*rows.Row, error) {
	return fn(rTx, tx, shard, f)
}

func NewAnd(fs ...Filter) Filter {
	if len(fs) == 0 {
		return Noop()
	}
	return filterFunc(func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (*rows.Row, error) {
		r := rows.NewRow()
		for i := range fs {
			x, err := fs[i].Apply(rTx, tx, shard, f)
			if err != nil {
				return nil, err
			}
			if x.IsEmpty() {
				return x, nil
			}
			if r.IsEmpty() {
				r = x
				continue
			}
			r = r.Intersect(x)
			if r.IsEmpty() {
				return r, nil
			}
		}
		return r, nil
	})
}

func Noop() Filter {
	return filterFunc(func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (*rows.Row, error) {
		return f, nil
	})
}

func Reject() Filter {
	return filterFunc(func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (*rows.Row, error) {
		return rows.NewRow(), nil
	})
}

func NewEq(name, value string) Filter {
	return filterFunc(func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (r *rows.Row, err error) {
		fd := newReadField(tx, []byte(name))
		id, ok := fd.get([]byte(value))
		if !ok {
			return rows.NewRow(), nil
		}
		r = rows.NewRow()
		err = cursor.Tx(rTx, name, func(c *rbf.Cursor) error {
			r, err = cursor.Row(c, shard, id)
			return err
		})
		if err != nil {
			return nil, err
		}
		return r.Intersect(f), nil
	})
}

func NewEqInt(name string, value int64) Filter {
	return filterFunc(func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (*rows.Row, error) {
		err := cursor.Tx(rTx, name, func(c *rbf.Cursor) error {
			var err error
			f, err = compare(c, shard, eq, value, 0, f)
			return err
		})
		if err != nil {
			return nil, err
		}
		return f, nil
	})
}

func NewNeq(name, value string) Filter {
	return filterFunc(func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (r *rows.Row, err error) {
		fd := newReadField(tx, []byte(name))
		id, ok := fd.get([]byte(value))
		if !ok {
			return f, nil
		}
		r = rows.NewRow()
		err = cursor.Tx(rTx, name, func(c *rbf.Cursor) error {
			r, err = cursor.Row(c, shard, id)
			return err
		})
		if err != nil {
			return nil, err
		}
		return r.Difference(f), nil
	})
}

func NewRe(name, value string) Filter {
	return filterFunc(func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (r *rows.Row, err error) {
		re, err := regexp.Compile(value)
		if err != nil {
			return nil, err
		}
		fd := newReadField(tx, []byte(name))

		match := fd.search(re)
		if len(match) == 0 {
			return rows.NewRow(), nil
		}
		if len(match) > 0 {
			err = cursor.Tx(rTx, name, func(c *rbf.Cursor) error {
				slices.Sort(match)
				fs := roaring.NewBitmapRowsUnion(match)
				err := c.ApplyFilter(0, fs)
				if err != nil {
					return err
				}
				o := rows.NewRowFromBitmap(fs.Results(shard))
				r = o.Intersect(f)
				return nil
			})
		}
		return
	})
}

func NewNre(name, value string) Filter {
	return filterFunc(func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, f *rows.Row) (r *rows.Row, err error) {
		re, err := regexp.Compile(value)
		if err != nil {
			return nil, err
		}
		fd := newReadField(tx, []byte(name))

		match := fd.search(re)
		if len(match) == 0 {
			return rows.NewRow(), nil
		}
		if len(match) > 0 {
			err = cursor.Tx(rTx, name, func(c *rbf.Cursor) error {
				slices.Sort(match)
				fs := roaring.NewBitmapRowsUnion(match)
				err := c.ApplyFilter(0, fs)
				if err != nil {
					return err
				}
				o := rows.NewRowFromBitmap(fs.Results(shard))
				r = o.Intersect(f)
				return nil
			})
		}
		return
	})
}
