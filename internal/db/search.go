package db

import (
	"errors"
	"strings"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/quantum"
)

type Query interface {
	View(ts time.Time) View
}

type View interface {
	Apply(tx *Tx) error
}

const layout = "20060102"

func (db *DB) Search(start, end time.Time, query Query) error {
	var views []string
	if date(start).Equal(date(end)) {
		views = []string{quantum.ViewByTimeUnit("std", start, 'D')}
	} else {
		views = quantum.ViewsByTimeRange("std", start, end, "D")
	}
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
					return qv.Apply(tx)
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

}
