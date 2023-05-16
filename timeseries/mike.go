package timeseries

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/cities"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/pkg/plot"
	"github.com/gernest/vince/pkg/timex"
	"github.com/gernest/vince/store"
	"github.com/gernest/vince/system"
	"github.com/gernest/vince/ua"
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

func (g *aggregate) Prop(el EntryList, by PROPS, f func(key string, sum *store.Sum) error) error {
	g.Reset()
	return el.EmitProp(g.u, g.s, by, &g.sum, f)
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

	// The first 16 bytes of ID and MetaKey are for user id and site id. We just copy it
	// directly from the buffer.
	copy(id[:], b.id[:])
	copy(meta[:], b.id[:])

	// Guarantee that aggregates are on per hour windows.
	ent.Emit(func(el EntryList) {
		defer group.Reset()
		group.Save(el)
		ts := time.Unix(el[0].Timestamp, 0)
		err := db.Update(func(txn *badger.Txn) error {
			id.Year(ts)
			return errors.Join(
				updateCalendar(ctx, txn, ts, id[:], &group.sum),
				updateMeta(ctx, txn, el, group, meta, ts),
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

func updateMeta(ctx context.Context, txn *badger.Txn, el EntryList, g *aggregate, x *MetaKey, ts time.Time) error {
	errs := make([]error, 0, PROPS_CITY)
	for i := 1; i <= int(PROPS_CITY); i++ {
		p := PROPS(i)
		errs = append(errs, p.Save(ctx, txn, el, g, x, ts))
	}
	return errors.Join(errs...)
}

func (p PROPS) Save(ctx context.Context, txn *badger.Txn, el EntryList, g *aggregate, x *MetaKey, ts time.Time) error {
	switch p {
	case PROPS_NAME, PROPS_PAGE, PROPS_ENTRY_PAGE, PROPS_EXIT_PAGE,
		PROPS_REFERRER, PROPS_UTM_MEDIUM, PROPS_UTM_SOURCE,
		PROPS_UTM_CAMPAIGN, PROPS_UTM_CONTENT, PROPS_UTM_TERM:
		return g.Prop(el, p, func(key string, sum *store.Sum) error {
			return updateCalendar(ctx, txn, ts, x.SetMeta(byte(p)).String(key), sum)
		})
		// properties from user agent
	case PROPS_UTM_DEVICE, PROPS_UTM_BROWSER, PROPS_BROWSER_VERSION, PROPS_OS, PROPS_OS_VERSION:
		return g.Prop(el, p, func(key string, sum *store.Sum) error {
			return updateCalendar(ctx, txn, ts, x.SetMeta(byte(p)).HashU16(ua.ToIndex(key)), sum)
		})
	case PROPS_COUNTRY, PROPS_REGION:
		return g.Prop(el, p, func(key string, sum *store.Sum) error {
			return updateCalendar(ctx, txn, ts, x.SetMeta(byte(p)).HashU16(cities.ToIndex(key)), sum)
		})
	case PROPS_CITY:
		return g.City(el, func(key uint32, sum *store.Sum) error {
			return updateCalendar(ctx, txn, ts, x.SetMeta(byte(p)).HashU32(key), sum)
		})
	default:
		return nil
	}
}

func (p PROPS) Key(ctx context.Context, key []byte) (s string) {
	switch p {
	case PROPS_NAME, PROPS_PAGE, PROPS_ENTRY_PAGE, PROPS_EXIT_PAGE,
		PROPS_REFERRER, PROPS_UTM_MEDIUM, PROPS_UTM_SOURCE,
		PROPS_UTM_CAMPAIGN, PROPS_UTM_CONTENT, PROPS_UTM_TERM:
		return string(key)
	case PROPS_UTM_DEVICE, PROPS_UTM_BROWSER, PROPS_BROWSER_VERSION, PROPS_OS, PROPS_OS_VERSION:
		return ua.FromIndex(binary.BigEndian.Uint16(key))
	case PROPS_COUNTRY, PROPS_REGION:
		return cities.NameFromIndex(binary.BigEndian.Uint16(key))
	case PROPS_CITY:
		err := GetGeo(ctx).View(func(txn *badger.Txn) error {
			x, err := txn.Get(key)
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					return nil
				}
				return err
			}
			return x.Value(func(val []byte) error {
				s = string(val)
				return nil
			})
		})
		if err != nil {
			log.Get(ctx).Err(err).
				Uint32("geoname_id", binary.BigEndian.Uint32(key)).
				Msg("failed to get city by geoname id")
		}
	}
	return
}

func updateCalendar(ctx context.Context, txn *badger.Txn, ts time.Time, key []byte, a *store.Sum) error {
	cache := caches.Calendar(ctx)
	hash := xxhash.Sum64(key)
	if x, ok := cache.Get(hash); ok {
		// The calendar was in cache we update it and save.
		cal := x.(*store.Calendar)
		a.UpdateCalendar(ts, cal)
		b, err := cal.Message().MarshalPacked()
		if err != nil {
			return fmt.Errorf("failed to marshal calendar %v", err)
		}
		return txn.Set(key, b)
	}
	x, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			// new entry we store a
			cal, err := store.ZeroCalendar(ts, a)
			if err != nil {
				return err
			}
			defer cache.Set(hash, cal, store.CacheCost)
			b, err := cal.Message().MarshalPacked()
			if err != nil {
				return fmt.Errorf("failed to marshal calendar %v", err)
			}
			return txn.Set(key, b)
		}
		return err
	}
	return x.Value(func(val []byte) error {
		msg, err := capnp.UnmarshalPacked(val)
		if err != nil {
			return err
		}
		cal, err := store.ReadRootCalendar(msg)
		if err != nil {
			return err
		}
		defer cache.Set(hash, cal, store.CacheCost)
		a.UpdateCalendar(ts, &cal)
		b, err := cal.Message().MarshalPacked()
		if err != nil {
			return fmt.Errorf("failed to marshal calendar %v", err)
		}
		return txn.Set(key, b)
	})
}

func ReadCalendars(ctx context.Context, pick timex.Range, uid, sid uint64) *plot.Data {
	start := time.Now()
	defer system.CalendarReadDuration.Observe(time.Since(start).Seconds())
	m := GetMike(ctx)
	id := newID()
	defer id.Release()
	id.SetUserID(uid)
	id.SetSiteID(sid)

	meta := newMetaKey()
	defer meta.Release()

	meta.SetUserID(uid)
	meta.SetSiteID(sid)
	var data plot.Data
	for _, r := range pick.Build() {
		ts := r.TS()
		id.Year(ts)
		meta.Year(ts)
		err := readCalendar(ctx, pick, m, id, meta, &data)
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to read calendar")
			return nil
		}
	}
	data.Timestamps = pick.Timestamps()
	return data.Build()
}

func readCalendar(ctx context.Context, pick timex.Range, m *badger.DB, id *ID, meta *MetaKey, data *plot.Data) error {
	return m.View(func(txn *badger.Txn) error {
		return errors.Join(
			readAllAggregate(txn, pick, id, data),
			readAllMeta(ctx, txn, pick, meta, data),
		)
	})
}

func readAllMeta(ctx context.Context, txn *badger.Txn, pick timex.Range, id *MetaKey, data *plot.Data) error {
	var errList []error
	for i := PROPS_NAME; i <= PROPS_CITY; i++ {
		id.SetMeta(byte(i))
		errList = append(errList, readMetaCal(txn, id.Prefix(), func(b []byte, c *store.Calendar) error {
			a, err := calToAggregate(pick, c)
			if err != nil {
				return err
			}
			data.Set(plot.Property(i-1), i.Key(ctx, b), a)
			return nil
		}))
	}
	return errors.Join(errList...)
}

func calToAggregate(pick timex.Range, c *store.Calendar) (o plot.AggregateValues, err error) {
	err = errors.Join(
		readAggregate(pick, c.SeriesVisitors, &o.Visitors),
		readAggregate(pick, c.SeriesViews, &o.Views),
		readAggregate(pick, c.SeriesEvents, &o.Events),
		readAggregate(pick, c.SeriesVisits, &o.Visits),
		readAggregate(pick, c.SeriesBounceRates, &o.BounceRate),
		readAggregate(pick, c.SeriesVisitDuration, &o.VisitDuration),
		readAggregate(pick, c.SeriesViewsPerVisit, &o.ViewsPerVisit),
	)
	return
}

func readAggregate(pick timex.Range, f func(time.Time, time.Time) ([]float64, error), o *[]float64) error {
	v, err := f(pick.From, pick.To)
	if err != nil {
		return err
	}
	*o = v
	return nil
}

func readMetaCal(txn *badger.Txn, prefix []byte, f func([]byte, *store.Calendar) error) error {
	o := badger.IteratorOptions{
		PrefetchValues: true,
		// be conservative. This should balance between high cardinality props with
		// low cardinality ones.
		PrefetchSize: 5,
		Prefix:       prefix,
	}
	it := txn.NewIterator(o)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {
		x := it.Item()
		key := x.Key()
		err := x.Value(func(val []byte) error {
			return store.CalendarFromBytes(val, func(c *store.Calendar) error {
				return f(key[len(prefix):], c)
			})
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func readAllAggregate(txn *badger.Txn, pick timex.Range, id *ID, data *plot.Data) error {
	data.All = &plot.Aggregate{
		Visitors:      &plot.Entry{},
		Views:         &plot.Entry{},
		Events:        &plot.Entry{},
		Visits:        &plot.Entry{},
		BounceRate:    &plot.Entry{},
		VisitDuration: &plot.Entry{},
		ViewsPerVisit: &plot.Entry{},
	}
	return readCal(txn, id[:], func(c *store.Calendar) error {
		return errors.Join(
			readEntry(pick, c.SeriesVisitors, &data.All.Visitors),
			readEntry(pick, c.SeriesViews, &data.All.Views),
			readEntry(pick, c.SeriesEvents, &data.All.Events),
			readEntry(pick, c.SeriesVisits, &data.All.Visits),
			readEntry(pick, c.SeriesBounceRates, &data.All.BounceRate),
			readEntry(pick, c.SeriesVisitDuration, &data.All.VisitDuration),
			readEntry(pick, c.SeriesViewsPerVisit, &data.All.ViewsPerVisit),
		)
	})
}

func readEntry(pick timex.Range, f func(time.Time, time.Time) ([]float64, error), e **plot.Entry) error {
	v, err := f(pick.From, pick.To)
	if err != nil {
		return err
	}
	*e = &plot.Entry{
		Values: v,
	}
	return nil
}

func readCal(txn *badger.Txn, key []byte, f func(*store.Calendar) error) error {
	it, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	}
	return it.Value(func(val []byte) error {
		return store.CalendarFromBytes(val, f)
	})
}
