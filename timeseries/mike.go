package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/system"
)

type Sum struct {
	Visitors      uint32
	Views         uint32
	Events        uint32
	Visits        uint32
	BounceRate    uint32
	VisitDuration uint32
	ViewsPerVisit uint32
}

func DropSite(ctx context.Context, uid, sid uint64) {
	start := time.Now()
	defer system.DropSiteDuration.Observe(time.Since(start).Seconds())

	db := GetMike(ctx)
	id := newMetaKey()
	defer id.Release()
	id.SetUserID(uid)
	id.SetSiteID(sid)
	// remove all keys under /user_id/site_id/ prefix.
	err := db.DropPrefix(id[:metricOffset])
	if err != nil {
		log.Get(ctx).Err(err).
			Uint64("uid", uid).
			Uint64("sid", sid).
			Msg("failed to delete site from stats storage")
	}

}

func Save(ctx context.Context, b *Buffer) {
	start := time.Now()
	defer system.SaveDuration.Observe(time.Since(start).Seconds())

	db := GetMike(ctx)
	meta := newMetaKey()

	defer func() {
		b.Release()
		meta.Release()
	}()

	svc := &saveContext{
		ls:  newMetaList(),
		idx: GetIndex(ctx).NewTransaction(true),
	}

	// Buffer.id has the same encoding as the first 16 bytes of meta. We just copy that
	// there is no need to re encode user id and site id.
	copy(meta[:], b.id[:])
	err := db.Update(func(txn *badger.Txn) error {
		svc.txn = txn
		return b.Build(ctx, func(p Property, key string, ts uint64, sum *Sum) error {
			return saveProperty(ctx, svc, meta, key, sum)
		})
	})
	if err != nil {
		// Transaction was discarded. We discard index transaction as well.
		svc.idx.Discard()
		log.Get(ctx).Err(err).Msg("failed to save ts buffer")
	} else {
		err := svc.idx.Commit()
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to commit to ts index")
		}
		svc.idx.Discard()
	}
}

func saveProperty(ctx context.Context,
	svc *saveContext,
	m *Key, text string, a *Sum) error {
	return errors.Join(
		savePropertyKey(ctx, svc, m.Metric(Visitors).Key(text), a.Visitors),
		savePropertyKey(ctx, svc, m.Metric(Views).Key(text), a.Views),
		savePropertyKey(ctx, svc, m.Metric(Events).Key(text), a.Events),
		savePropertyKey(ctx, svc, m.Metric(Visits).Key(text), a.Visits),
		savePropertyKey(ctx, svc, m.Metric(BounceRate).Key(text), a.BounceRate),
		savePropertyKey(ctx, svc, m.Metric(VisitDuration).Key(text), a.VisitDuration),
		savePropertyKey(ctx, svc, m.Metric(ViewsPerVisit).Key(text), a.ViewsPerVisit),
	)
}

type saveContext struct {
	txn *badger.Txn
	ls  *metaList
	idx *badger.Txn
}

func (ctx *saveContext) saveIndex(key *bytes.Buffer) error {
	ctx.ls.ls = append(ctx.ls.ls, key)
	return ctx.idx.Set(key.Bytes(), []byte{})
}

// Transaction keep reference to keys. We need this to make sure we reuse ID and properly
// release them back to the pool when done.
type metaList struct {
	ls []*bytes.Buffer
}

func newMetaList() *metaList {
	return metaKeyPool.New().(*metaList)
}

func (ls *metaList) Get() *bytes.Buffer {
	b := smallBufferpool.Get().(*bytes.Buffer)
	ls.ls = append(ls.ls, b)
	return b
}

func (ls *metaList) Reset() {
	for _, v := range ls.ls {
		v.Reset()
		smallBufferpool.Put(v)
	}
	ls.ls = ls.ls[:0]
}

func (ls *metaList) Release() {
	ls.Reset()
	metaListPool.Put(ls)
}

var metaListPool = &sync.Pool{
	New: func() any {
		return &metaList{
			ls: make([]*bytes.Buffer, 0, 1<<10),
		}
	},
}

var smallBufferpool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func savePropertyKey(ctx context.Context, svc *saveContext, id *IDToSave, a uint32) error {
	svc.ls.ls = append(svc.ls.ls, id.mike, id.index)
	txn := svc.txn
	key := id.mike.Bytes()
	x, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			b := make([]byte, 8)
			binary.BigEndian.PutUint64(b, math.Float64bits(float64(a)))
			err := txn.Set(key, b)
			if err != nil {
				return err
			}
			// We have successfully set the key, now we set the index. We only index
			// new keys. Since keys are stable, there is no need to index again when
			// doing update.
			err = svc.saveIndex(id.index)
			if err != nil {
				log.Get(ctx).Err(err).
					Msg("failed to save index")
			}
			return err
		}
		return err
	}
	var read float64
	x.Value(func(val []byte) error {
		read = math.Float64frombits(
			binary.BigEndian.Uint64(val),
		)
		return nil
	})
	read += float64(a)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, math.Float64bits(read))
	return txn.Set(key, b)
}
