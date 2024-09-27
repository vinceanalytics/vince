package ro2

import (
	"errors"
	"time"

	"github.com/gernest/roaring/shardwidth"
	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/bsi"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/cursor"
	"github.com/vinceanalytics/vince/internal/rbf/quantum"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

const (
	domainField = "domain"
	idField     = "id"
)

func (o *Store) CurrentVisitors(domain string) (visitors uint64, err error) {
	end := time.Now().UTC().Truncate(time.Minute)
	start := end.Add(-5 * time.Minute)

	r := roaring64.New()

	err = o.View(func(tx *Tx) error {
		domainId, ok := tx.ID(uint64(alicia.DOMAIN), domain)
		if !ok {
			return nil
		}
		shard := tx.Seq() / shardwidth.ShardWidth

		// only search through the current shard for current visitors
		return o.shards.View(shard, func(rtx *rbf.Tx) error {
			f := quantum.NewField()
			defer f.Release()
			return f.Minute("domain", start, end, func(b []byte) error {
				return viewCu(rtx, string(b), func(rCu *rbf.Cursor) error {
					dRow, err := cursor.Row(rCu, shard, domainId)
					if err != nil {
						return err
					}
					if dRow.IsEmpty() {
						return nil
					}
					return viewCu(rtx, "id"+string(b[len(domainField):]), func(rCu *rbf.Cursor) error {
						return bsi.Extract(rCu, shard, dRow, func(column uint64, value int64) {
							r.Add(uint64(value))
						})
					})
				})
			})
		})
	})
	visitors = r.GetCardinality()
	return
}

func viewCu(rtx *rbf.Tx, name string, f func(rCu *rbf.Cursor) error) error {
	cu, err := rtx.Cursor(name)
	if err != nil {
		if errors.Is(err, rbf.ErrBitmapNotFound) {
			return nil
		}
		return err
	}
	defer cu.Close()
	return f(cu)
}

func (o *Store) Visitors(domain string) (visitors uint64, err error) {
	r := roaring64.New()
	err = o.View(func(tx *Tx) error {
		domainId, ok := tx.ID(uint64(alicia.DOMAIN), domain)
		if !ok {
			return nil
		}
		shards := o.Shards(tx)

		for i := range shards {
			shard := shards[i]
			err := o.shards.View(shard, func(rtx *rbf.Tx) error {
				return viewCu(rtx, domainField, func(rCu *rbf.Cursor) error {
					dRow, err := cursor.Row(rCu, shard, domainId)
					if err != nil {
						return err
					}
					if dRow.IsEmpty() {
						return nil
					}
					return viewCu(rtx, idField, func(rCu *rbf.Cursor) error {
						return bsi.Extract(rCu, shard, dRow, func(column uint64, value int64) {
							r.Add(uint64(value))
						})
					})
				})
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	visitors = r.GetCardinality()
	return
}
