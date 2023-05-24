package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
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

type aggregate struct {
	u, s *roaring64.Bitmap
	sum  Sum
}

var groupPool = &sync.Pool{
	New: func() any {
		return &aggregate{
			u: roaring64.New(),
			s: roaring64.New(),
		}
	},
}

func (g *aggregate) Save(el EntryList) {
	el.Count(g.u, g.s, &g.sum)
}

func (g *aggregate) Prop(ctx context.Context, cf *saveContext, el EntryList, by Property, f func(key string, sum *Sum) error) error {
	return el.EmitProp(ctx, cf, g.u, g.s, by, &g.sum, f)
}

func (g *aggregate) Release() {
	groupPool.Put(g)
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
	err := db.DropPrefix(id[:aggregateTypeOffset])
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
	defer b.Release()
	group := groupPool.Get().(*aggregate)
	ent := EntryList(b.entries)
	meta := newMetaKey()
	defer meta.Release()
	mls := newMetaList()
	defer mls.Release()
	sctx := &saveContext{}
	sctx.Reset(ctx)
	defer func() {
		sctx.Release()
	}()
	// Buffer.id has the same encoding as the first 16 bytes of meta. We just copy that
	// there is no need to re encode user id and site id.
	copy(meta[:], b.id[:])

	// Guarantee that aggregates are on per hour windows.
	ent.Emit(func(el EntryList) {
		// This ensures indexes are committed and resources reclaimed for reuse.
		defer sctx.Reset(ctx)

		group.Save(el)
		ts := time.Unix(el[0].Timestamp, 0)
		meta.Timestamp(ts)
		err := db.Update(func(txn *badger.Txn) error {
			sctx.txn = txn
			return errors.Join(
				saveProp(ctx, sctx, meta.SetProp(Base), "__root__", &group.sum),
				updateMeta(ctx, sctx, el, group, meta, ts),
			)
		})
		if err != nil {
			log.Get(ctx).Err(err).
				Uint64("uid", b.UID()).
				Uint64("sid", b.SID()).
				Msg("failed to save hourly stats")
		}
	})
}

func updateMeta(ctx context.Context, sctx *saveContext, el EntryList, g *aggregate, x *Key, ts time.Time) error {
	errs := make([]error, 0, City)
	for i := 1; i <= int(City); i++ {
		p := Property(i)
		errs = append(errs, p.Save(ctx, sctx, el, g, x, ts))
	}
	return errors.Join(errs...)
}

func (p Property) Save(ctx context.Context, sctx *saveContext, el EntryList, g *aggregate, x *Key, ts time.Time) error {
	return g.Prop(ctx, sctx, el, p, func(key string, sum *Sum) error {
		return saveProp(ctx, sctx, x.SetProp(p), key, sum)
	})
}

func saveProp(ctx context.Context,
	sctx *saveContext,
	m *Key, text string, a *Sum) error {
	return errors.Join(
		updateMetaKey(ctx, sctx, m.SetAggregateType(Visitors).Key(text), a.Visitors),
		updateMetaKey(ctx, sctx, m.SetAggregateType(Views).Key(text), a.Views),
		updateMetaKey(ctx, sctx, m.SetAggregateType(Events).Key(text), a.Events),
		updateMetaKey(ctx, sctx, m.SetAggregateType(Visits).Key(text), a.Visits),
		updateMetaKey(ctx, sctx, m.SetAggregateType(BounceRate).Key(text), a.BounceRate),
		updateMetaKey(ctx, sctx, m.SetAggregateType(VisitDuration).Key(text), a.VisitDuration),
		updateMetaKey(ctx, sctx, m.SetAggregateType(ViewsPerVisit).Key(text), a.ViewsPerVisit),
	)
}

type saveContext struct {
	txn *badger.Txn
	ls  *metaList
	idx *badger.Txn
}

func (ctx *saveContext) Reset(rctx context.Context) {
	if ctx.ls != nil {
		ctx.ls.Reset()
	} else {
		ctx.ls = newMetaList()
	}
	if ctx.idx != nil {
		ctx.idx.Commit()
		ctx.idx.Discard()
	} else {
		ctx.idx = GetIndex(rctx).NewTransaction(true)
	}
}

func (ctx *saveContext) Release() {
	ctx.idx.Discard()
	ctx.ls.Release()
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

func updateKeyRaw(ctx context.Context, sctx *saveContext, id *IDToSave, a uint32) error {
	txn := sctx.txn
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
			err = sctx.saveIndex(id.index)
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

func updateMetaKey(ctx context.Context, sctx *saveContext, id *IDToSave, a uint32) error {
	sctx.ls.ls = append(sctx.ls.ls, id.mike, id.index)
	return updateKeyRaw(ctx, sctx, id, a)
}
