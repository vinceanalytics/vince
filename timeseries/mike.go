package timeseries

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/store"
	"google.golang.org/protobuf/proto"
)

var commonHash = &sync.Map{}

var commonProps = []string{
	"pageviews",
	// devices
	"mobile",
	"tablet",
	"laptop",
	"desktop",
}

var commonKeysSet = roaring64.NewBitmap()

func init() {
	for _, h := range commonProps {
		x := hashKey(h)
		commonHash.Store(x, h)
		commonKeysSet.Add(x)
	}
}

func keyIsCommon(h uint64) bool {
	return commonKeysSet.Contains(h)
}

// case insensitive hash of key
func hashKey(key string) uint64 {
	return xxhash.Sum64String(strings.ToLower(key))
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
			{
				key := id.Year(ts).SetTable(byte(TABLE_YEAR))[:]
				err := updateCalendar(txn, ts, key, &group.sum)
				if err != nil {
					return err
				}
			}
			{
				err := saveProps(txn, el, group, meta, PROPS_NAME, ts)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to save hourly stats ")
		}
	})
}

func saveProps(txn *badger.Txn, el EntryList, g *Group, x *MetaKey, by PROPS, ts time.Time) error {
	var m Hash
	err := updateProp(txn, el, g, x, by, ts, &m)
	if err != nil {
		return err
	}
	if len(m.Hash) == 0 {
		// avoid touching the index for common keys
		return nil
	}
	return updatePropHash(txn, x, by, ts, &m)
}

func updateProp(txn *badger.Txn, el EntryList, g *Group, x *MetaKey, by PROPS, ts time.Time, m *Hash) error {
	x.Year(ts).SetTable(byte(TABLE_YEAR)).SetMeta(byte(by))
	return g.Prop(el, by, func(key string, sum *store.Sum) error {
		h := hashKey(key)
		if m.Hash == nil {
			m.Hash = make(map[uint64]string)
		}
		if _, ok := m.Hash[h]; !ok {
			if !keyIsCommon(h) {
				m.Hash[h] = key
			}
		}
		x.SetHash(h)
		return updateCalendar(txn, ts, x[:], sum)
	})
}

func updatePropHash(txn *badger.Txn, x *MetaKey, by PROPS, ts time.Time, m *Hash) error {
	key := x.Year(ts).SetTable(byte(TABLE_HASH)).SetMeta(byte(by))[:hashOffset]
	it, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			b, err := proto.Marshal(m)
			if err != nil {
				return err
			}
			return txn.Set(key, b)
		}
		return err
	}
	return it.Value(func(val []byte) error {
		var o Hash
		err := proto.Unmarshal(val, &o)
		if err != nil {
			return err
		}
		proto.Merge(&o, m)
		b, err := proto.Marshal(&o)
		if err != nil {
			return err
		}
		return txn.Set(key, b)
	})
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
