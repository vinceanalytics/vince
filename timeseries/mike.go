package timeseries

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/store"
)

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
	visitors, visits, events := el.Count(g.u, g.s)
	g.sum.SetVisitors(visitors)
	g.sum.SetVisits(visits)
	g.sum.SetEvents(events)
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
	id.SetSiteID(b.SID())
	id.SetUserID(b.UID())
	ent.Emit(func(el EntryList) {
		defer group.Reset()
		group.Save(el)
		ts := time.Unix(el[0].Timestamp, 0)
		err := db.Update(func(txn *badger.Txn) error {
			key := id.Year(ts).SetTable(byte(TABLE_YEAR))[:]
			return updateCalendar(txn, ts, key, group.sum)
		})
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to save hourly stats ")
		}
	})
}

// creates a new calendar for ts year and updates the sum of this date. For existing
// calendar we just update the sums for the date.
func updateCalendar(txn *badger.Txn, ts time.Time, key []byte, a store.Sum) error {
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
