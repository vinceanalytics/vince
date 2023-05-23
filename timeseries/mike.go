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
	"github.com/gernest/vince/store"
	"github.com/gernest/vince/system"
)

type aggregate struct {
	u, s *roaring64.Bitmap
	sum  store.Sum
}

var groupPool = &sync.Pool{
	New: func() any {
		return &aggregate{
			u: roaring64.New(),
			s: roaring64.New(),
		}
	},
}

func (g *aggregate) Reset() {
	g.u.Clear()
	g.s.Clear()
	g.sum = store.Sum{}
}

func (g *aggregate) Save(el EntryList) {
	g.Reset()
	el.Count(g.u, g.s, &g.sum)
}

func (g *aggregate) Prop(ctx context.Context, cf *CityFinder, el EntryList, by PROPS, f func(key string, sum *store.Sum) error) error {
	g.Reset()
	return el.EmitProp(ctx, cf, g.u, g.s, by, &g.sum, f)
}

func (g *aggregate) City(el EntryList, f func(key uint32, sum *store.Sum) error) error {
	g.Reset()
	return el.EmitCity(g.u, g.s, &g.sum, f)
}

func (g *aggregate) Release() {
	g.Reset()
	groupPool.Put(g)
}

func DropSite(ctx context.Context, uid, sid uint64) {
	start := time.Now()
	defer system.DropSiteDuration.Observe(time.Since(start).Seconds())

	db := GetMike(ctx)
	id := newID()
	defer id.Release()

	id.SetUserID(uid)
	id.SetSiteID(sid)

	err := db.Update(func(txn *badger.Txn) error {
		o := badger.IteratorOptions{
			Prefix: id.SitePrefix(),
		}
		it := txn.NewIterator(o)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			err := txn.Delete(it.Item().KeyCopy(nil))
			if err != nil {
				return err
			}
		}
		return nil
	})
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
	id := newID()
	defer id.Release()
	meta := newMetaKey()
	defer meta.Release()
	ls := newIDList()
	defer ls.Release()
	mls := newMetaList()
	defer mls.Release()

	// The first 16 bytes of ID and MetaKey are for user id and site id. We just copy it
	// directly from the buffer.
	copy(id[:], b.id[:])
	copy(meta[:], b.id[:])

	// Guarantee that aggregates are on per hour windows.
	ent.Emit(func(el EntryList) {
		defer func() {
			group.Reset()
			ls.Reset()
		}()
		group.Save(el)
		ts := time.Unix(el[0].Timestamp, 0)

		err := db.Update(func(txn *badger.Txn) error {
			id.Timestamp(ts)
			return errors.Join(
				updateRoot(ctx, ls, txn, ts, id, &group.sum),
				updateMeta(ctx, mls, txn, el, group, meta, ts),
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

func updateMeta(ctx context.Context, ls *metaList, txn *badger.Txn, el EntryList, g *aggregate, x *MetaKey, ts time.Time) error {
	f := newCityFinder(ctx)
	defer f.Release()
	errs := make([]error, 0, PROPS_city)
	for i := 1; i <= int(PROPS_city); i++ {
		p := PROPS(i)
		errs = append(errs, p.Save(ctx, f, ls, txn, el, g, x, ts))
	}
	return errors.Join(errs...)
}

func (p PROPS) Save(ctx context.Context, f *CityFinder, ls *metaList, txn *badger.Txn, el EntryList, g *aggregate, x *MetaKey, ts time.Time) error {
	return g.Prop(ctx, f, el, p, func(key string, sum *store.Sum) error {
		if key == "" {
			// skip empty keys
			return nil
		}
		return updateCalendarText(ctx, ls, txn, ts, x.SetProp(byte(p)), key, sum)
	})
}

type CityFinder struct {
	txn *badger.Txn
	key [4]byte
}

func (c *CityFinder) Release() {
	c.txn.Discard()
}

func (c *CityFinder) Get(ctx context.Context, geoname uint32) (s string) {
	binary.BigEndian.PutUint32(c.key[:], geoname)
	x, err := c.txn.Get(c.key[:])
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed to get city by geoname id")
		return ""
	}
	x.Value(func(val []byte) error {
		s = string(val)
		return nil
	})
	return
}

func newCityFinder(ctx context.Context) *CityFinder {
	return &CityFinder{
		txn: GetGeo(ctx).NewTransaction(false),
	}
}

func updateCalendarText(ctx context.Context,
	ls *metaList,
	txn *badger.Txn, ts time.Time,
	m *MetaKey, text string, a *store.Sum) error {
	return errors.Join(
		updateMetaKey(ctx, ls, txn, m.SetAggregateType(METRIC_TYPE_visitors).String(text), a.Visitors),
		updateMetaKey(ctx, ls, txn, m.SetAggregateType(METRIC_TYPE_views).String(text), a.Views),
		updateMetaKey(ctx, ls, txn, m.SetAggregateType(METRIC_TYPE_events).String(text), a.Events),
		updateMetaKey(ctx, ls, txn, m.SetAggregateType(METRIC_TYPE_visits).String(text), a.Visits),
		updateMetaKey(ctx, ls, txn, m.SetAggregateType(METRIC_TYPE_bounce_rate).String(text), a.BounceRate),
		updateMetaKey(ctx, ls, txn, m.SetAggregateType(METRIC_TYPE_visitDuration).String(text), a.VisitDuration),
		updateMetaKey(ctx, ls, txn, m.SetAggregateType(METRIC_TYPE_viewsPerVisit).String(text), a.ViewsPerVisit),
	)
}

func updateRoot(ctx context.Context, ls *idList, txn *badger.Txn, ts time.Time, id *ID, a *store.Sum) error {
	return errors.Join(
		updateKey(ctx, ls, METRIC_TYPE_visitors, txn, id, a.Visitors),
		updateKey(ctx, ls, METRIC_TYPE_views, txn, id, a.Views),
		updateKey(ctx, ls, METRIC_TYPE_events, txn, id, a.Events),
		updateKey(ctx, ls, METRIC_TYPE_visits, txn, id, a.Visits),
		updateKey(ctx, ls, METRIC_TYPE_bounce_rate, txn, id, a.BounceRate),
		updateKey(ctx, ls, METRIC_TYPE_visitDuration, txn, id, a.VisitDuration),
		updateKey(ctx, ls, METRIC_TYPE_viewsPerVisit, txn, id, a.ViewsPerVisit),
	)
}

// Transaction keep reference to keys. We need this to make sure we reuse ID and properly
// release them back to the pool when done.
type idList struct {
	ls []*ID
}

func newIDList() *idList {
	return idListPool.New().(*idList)
}

func (ls *idList) Reset() {
	for _, v := range ls.ls {
		v.Release()
	}
	ls.ls = ls.ls[:0]
}

func (ls *idList) Release() {
	ls.Reset()
	idListPool.Put(ls)
}

var idListPool = &sync.Pool{
	New: func() any {
		return &idList{
			ls: make([]*ID, 0, 1<<10),
		}
	},
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

func updateKey(ctx context.Context, ls *idList, kind METRIC_TYPE, txn *badger.Txn, id *ID, a uint32) error {
	clone := id.Clone().SetAggregateType(kind)
	ls.ls = append(ls.ls, clone)
	key := clone[:]
	return updateKeyRaw(ctx, txn, key, a)
}

func updateKeyRaw(ctx context.Context, txn *badger.Txn, key []byte, a uint32) error {
	x, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			b := make([]byte, 8)
			binary.BigEndian.PutUint64(b, math.Float64bits(float64(a)))
			return txn.Set(key, b)
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

func updateMetaKey(ctx context.Context, ls *metaList, txn *badger.Txn, id *bytes.Buffer, a uint32) error {
	ls.ls = append(ls.ls, id)
	key := id.Bytes()
	return updateKeyRaw(ctx, txn, key, a)
}
