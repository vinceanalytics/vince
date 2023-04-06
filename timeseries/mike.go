package timeseries

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/store"
	"github.com/gernest/vince/ua"
)

func hashKey(key string) uint64 {
	return xxhash.Sum64String(key)
}

type Group struct {
	u, s *roaring64.Bitmap
	sum  store.Sum
}

var groupPool = &sync.Pool{
	New: func() any {
		return &Group{
			u:   roaring64.New(),
			s:   roaring64.New(),
			sum: store.ZeroSum(),
		}
	},
}

func (g *Group) Reset() {
	g.u.Clear()
	g.s.Clear()
	g.sum.Reuse()
}

func (g *Group) Save(el EntryList) {
	g.Reset()
	g.sum.SetValues(el.Count(g.u, g.s))
}

func (g *Group) Prop(el EntryList, by PROPS, f func(key string, sum *store.Sum) error) error {
	g.Reset()
	return el.EmitProp(g.u, g.s, by, &g.sum, f)
}

func (g *Group) Release() {
	g.Reset()
	groupPool.Put(g)
}

func Save(ctx context.Context, b *Buffer) {
	db := GetMike(ctx)
	defer b.Release()
	group := groupPool.Get().(*Group)
	ent := EntryList(b.entries)
	id := newID()
	defer id.Release()
	meta := newMetaKey()
	defer meta.Release()

	sid := b.SID()
	uid := b.UID()

	id.SetSiteID(sid)
	id.SetUserID(uid)
	meta.SetSiteID(sid)
	meta.SetUserID(uid)

	ent.Emit(func(el EntryList) {
		defer group.Reset()
		group.Save(el)
		ts := time.Unix(el[0].Timestamp, 0)
		err := db.Update(func(txn *badger.Txn) error {
			id.Year(ts).SetTable(byte(TABLE_YEAR))
			return errors.Join(
				updateCalendar(txn, ts, id[:], &group.sum),
				updateFromUA(txn, el, group, meta, ts),
			)
		})
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to save hourly stats ")
		}
	})
}

// compute and update calendars for values derived from user agent.
func updateFromUA(txn *badger.Txn, el EntryList, g *Group, x *MetaKey, ts time.Time) error {
	x.Year(ts).SetTable(byte(TABLE_YEAR)).SetMeta(byte(PROPS_UTM_DEVICE))
	return errors.Join(
		g.Prop(el, PROPS_UTM_DEVICE, func(key string, sum *store.Sum) error {
			return updateCalendar(txn, ts, x.HashU16(ua.ToIndex(key)), sum)
		}),
		g.Prop(el, PROPS_UTM_BROWSER, func(key string, sum *store.Sum) error {
			return updateCalendar(txn, ts, x.HashU16(ua.ToIndex(key)), sum)
		}),
		g.Prop(el, PROPS_BROWSER_VERSION, func(key string, sum *store.Sum) error {
			return updateCalendar(txn, ts, x.HashU16(ua.ToIndex(key)), sum)
		}),
		g.Prop(el, PROPS_OS, func(key string, sum *store.Sum) error {
			return updateCalendar(txn, ts, x.HashU16(ua.ToIndex(key)), sum)
		}),
		g.Prop(el, PROPS_OS_VERSION, func(key string, sum *store.Sum) error {
			return updateCalendar(txn, ts, x.HashU16(ua.ToIndex(key)), sum)
		}),
	)
}

// creates a new calendar for ts year and updates the sum of this date. For existing
// calendar we just update the sums for the date.
func updateCalendar(txn *badger.Txn, ts time.Time, key []byte, a *store.Sum) error {
	x, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			// new entry we store a
			cal, err := store.ZeroCalendar(ts, a)
			if err != nil {
				return err
			}
			defer cal.Message().Release()
			b, err := cal.Message().MarshalPacked()
			if err != nil {
				return fmt.Errorf("failed to marshal calendar %v", err)
			}
			return txn.Set(key, b)
		}
		return err
	}
	return x.Value(func(val []byte) error {
		cal, err := store.CalendarFromBytes(val)
		if err != nil {
			return err
		}
		defer cal.Message().Release()
		cal.Update(ts, a)
		b, err := cal.Message().MarshalPacked()
		if err != nil {
			return fmt.Errorf("failed to marshal calendar %v", err)
		}
		return txn.Set(key, b)
	})
}

func (x PROPS) ShouldIndex() bool {
	switch x {
	case PROPS_COUNTRY,
		PROPS_REGION,
		PROPS_UTM_DEVICE,
		PROPS_UTM_BROWSER,
		PROPS_BROWSER_VERSION,
		PROPS_OS, PROPS_OS_VERSION:
		return false
	default:
		return true
	}
}
