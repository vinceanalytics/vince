package db

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl"
	"github.com/gernest/rbf/quantum"
	"github.com/gernest/rows"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

type Query interface {
	View(ts time.Time) View
}

type View interface {
	Apply(tx *Tx, columns *rows.Row) error
}

const layout = "20060102"

func (db *DB) Search(start, end time.Time, filters []*v1.Filter, query Query) error {
	var views []string
	if date(start).Equal(date(end)) {
		views = []string{quantum.ViewByTimeUnit("", start, 'D')}
	} else {
		views = quantum.ViewsByTimeRange("", start, end, "D")
	}
	tsFIlter := filterTime(start, end)
	fs := filterProperties(filters)

	return db.View(func(tx *Tx) error {
		for _, view := range views {
			it, err := tx.Txn.Get([]byte(view))
			if err != nil {
				if !errors.Is(err, badger.ErrKeyNotFound) {
					return err
				}
				continue
			}
			var shards []uint64
			err = it.Value(func(val []byte) error {
				shards = arrow.Uint64Traits.CastFromBytes(val)
				return nil
			})
			if err != nil {
				return err
			}
			ts, err := time.Parse(layout, strings.TrimPrefix(view, "std_"))
			if err != nil {
				return err
			}
			qv := query.View(ts)
			for _, shard := range shards {
				err = tx.DB.ViewShard(shard, func(itx *rbf.Tx) error {
					tx.View = view
					tx.Shard = shard
					tx.Tx = itx

					f, err := tsFIlter(tx)
					if err != nil {
						return err
					}
					if f.IsEmpty() {
						return nil
					}
					ps, err := fs(tx, f)
					if err != nil {
						return err
					}
					if ps.IsEmpty() {
						return nil
					}
					return qv.Apply(tx, ps)
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

}

func filterTime(start, end time.Time) func(tx *Tx) (*rows.Row, error) {
	from := uint64(start.UTC().UnixMilli())
	to := uint64(end.UTC().UnixMilli())
	b := new(ViewFmt)
	return func(tx *Tx) (*rows.Row, error) {
		view := b.Format(tx.View, "ts")
		return dsl.CompareValueBSI(tx.Tx, view, tx.Shard,
			dsl.RANGE,
			from, to, nil)
	}
}

type Filter func(tx *Tx, columns *rows.Row) (*rows.Row, error)

func noop(_ *Tx, _ *rows.Row) (*rows.Row, error) {
	return rows.NewRow(), nil
}

func filterProperties(fs []*v1.Filter) Filter {
	if len(fs) == 0 {
		return noop
	}
	ls := make([]Filter, len(fs))
	for i := range fs {
		ls[i] = filterProperty(fs[i])
	}
	return func(tx *Tx, filter *rows.Row) (*rows.Row, error) {
		r := rows.NewRow()
		for i := range ls {
			x, err := ls[i](tx, filter)
			if err != nil {
				return nil, err
			}
			if i == 0 {
				r = x
			} else {
				r = r.Intersect(x)
			}
			if r.IsEmpty() {
				return r, nil
			}
		}
		return r, nil
	}
}

func filterProperty(f *v1.Filter) Filter {
	switch f.Op {
	case v1.Filter_equal, v1.Filter_not_equal:
		var id uint64
		var seen bool
		var once sync.Once
		var b ViewFmt
		return func(tx *Tx, filter *rows.Row) (*rows.Row, error) {
			once.Do(func() {
				id, seen = tx.find(f.Property.String(), []byte(f.Value))
			})
			if !seen {
				return rows.NewRow(), nil
			}
			view := b.Format(tx.View, f.Property.String())
			switch f.Op {
			case v1.Filter_equal:
				return dsl.CompareValueBSI(tx.Tx, view, tx.Shard, dsl.EQ, id, 0, filter)
			case v1.Filter_not_equal:
				return dsl.CompareValueBSI(tx.Tx, view, tx.Shard, dsl.NEQ, id, 0, filter)
			default:
				return rows.NewRow(), nil
			}
		}
	default:
		return noop
	}
}
