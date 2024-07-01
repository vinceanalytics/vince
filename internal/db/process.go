package db

import (
	"context"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/schema"
	"github.com/gernest/rbf/quantum"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/logger"
)

func (db *DB) Process(ctx context.Context, events <-chan *v1.Data) {
	tr := &Translator{db: db.db, update: true}
	defer tr.Release()

	seq, err := db.db.GetSequence(seqPrefix, 1<<30)
	if err != nil {
		logger.Fail("initializing sequence", "err", err)
	}
	defer seq.Release()

	ts := time.NewTicker(time.Minute)
	defer ts.Stop()
	sx, err := schema.NewSchema[*v1.Data](tr)
	if err != nil {
		logger.Fail("initializing batch schema", "err", err)
	}
	currentShard := ^uint64(0)
	begin := true

	defer func() {
		err = db.save(tr, sx, currentShard, time.Now())
		if err != nil {
			logger.Fail("saving events batch", "err", err)
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-events:
			next, err := seq.Next()
			if err != nil {
				logger.Fail("generating sequence", "err", err)
			}
			shard := next / shardwidth.ShardWidth
			if shard != currentShard {
				if !begin {
					// changed shard, save previous shard
					err = db.save(tr, sx, currentShard, time.Now())
					if err != nil {
						logger.Fail("saving events batch", "err", err)
					}
				} else {
					begin = false
				}
				currentShard = shard
			}
			err = sx.Write(next, e)
			if err != nil {
				logger.Fail("writing event batch", "err", err)
			}
		case now := <-ts.C:
			if !begin {
				err = db.save(tr, sx, currentShard, now)
				if err != nil {
					logger.Fail("saving events batch", "err", err)
				}
			}
		}
	}
}

func (db *DB) save(tr *Translator, batch *schema.Schema[*v1.Data], shard uint64, ts time.Time) error {
	defer batch.Reset()
	err := tr.Commit()
	if err != nil {
		return err
	}
	view := quantum.ViewByTimeUnit("", ts, 'D')
	vfmt := ViewFmt{}
	err = db.shards.Update(shard, func(tx *rbf.Tx) error {
		return batch.Range(func(name string, r *roaring.Bitmap) error {
			_, err := tx.AddRoaring(vfmt.Format(view, name), r)
			return err
		})
	})
	if err != nil {
		return err
	}
	// update view/shard mapping
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(vfmt.Shard(view, shard)), []byte{})
	})
}
