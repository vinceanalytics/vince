package timeseries

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/dgraph-io/badger/v3"
	"github.com/gernest/vince/log"
	"google.golang.org/protobuf/proto"
)

// Mike is the permanent storage for the events data. Data stored here is aggregated
// and broken down. All data is still stored in parquet format. This  only supports
// reads and writes, nothing is ever deleted from this storage.
type Mike struct {
	db *badger.DB
}

type Group struct {
	u, s  *roaring64.Bitmap
	props [PROPS_CITY]EntrySegment
}

var groupPool = &sync.Pool{
	New: func() any {
		return &Group{
			u: roaring64.New(),
			s: roaring64.New(),
		}
	},
}

func (g *Group) Reset() {
	for i := range g.props {
		g.props[i].Reuse()
	}
	g.u.Clear()
	g.s.Clear()
}

func (g *Group) Release() {
	g.Reset()
	groupPool.Put(g)
}

func (m *Mike) Save(ctx context.Context, b *Buffer, uid, sid uint64) {
	defer b.Release()
	group := groupPool.Get().(*Group)

	ent := EntryList(b.entries)
	id := NewID()
	defer id.Release()
	id.SetSiteID(sid)
	id.SetUserID(uid)
	a := &Aggr{}

	ent.Emit(func(i int, el EntryList) {
		defer group.Reset()
		x := &group.props[i]
		x.Save(el...)
		a.Reset()
		a.Total = el.Aggr(group.u, group.s)
		x.Aggr(group.u, group.s, a)
		ts := time.Unix(el[0].Timestamp, 0)
		err := m.db.Update(func(txn *badger.Txn) error {
			err := updateRoot(txn, ts, id, a.Total)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_NAME, a.Name)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_PAGE, a.Pathname)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_ENTRY_PAGE, a.EntryPage)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_EXIT_PAGE, a.ExitPage)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_REFERRER, a.Referrer)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_UTM_MEDIUM, a.UtmMedium)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_UTM_SOURCE, a.UtmSource)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_UTM_CAMPAIGN, a.UtmCampaign)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_UTM_CONTENT, a.UtmContent)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_UTM_TERM, a.UtmTerm)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_UTM_DEVICE, a.ScreenSize)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_UTM_BROWSER, a.Browser)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_BROWSER_VERSION, a.BrowserVersion)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_OS, a.OperatingSystem)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_OS_VERSION, a.OperatingSystemVersion)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_COUNTRY, a.CountryCode)
			if err != nil {
				return err
			}
			err = updateMeta(txn, ts, id, PROPS_COUNTRY, a.Region)
			if err != nil {
				return err
			}
			return updateMeta(txn, ts, id, PROPS_CITY, a.CityGeoNameId)
		})
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to save hourly stats ")
		}
	})
}

func (e *EntrySegment) Save(entries ...*Entry) {
	for _, n := range entries {
		e.Name = append(e.Name, n)
		e.Pathname = append(e.Pathname, n)
		e.EntryPage = append(e.EntryPage, n)
		e.ExitPage = append(e.ExitPage, n)
		if n.Referrer != "" {
			e.Referrer = append(e.Referrer, n)
		}
		if n.ReferrerSource != "" {
			e.ReferrerSource = append(e.ReferrerSource, n)
		}
		if n.UtmMedium != "" {
			e.UtmMedium = append(e.UtmMedium, n)
		}
		if n.UtmSource != "" {
			e.UtmSource = append(e.UtmSource, n)
		}
		if n.UtmCampaign != "" {
			e.UtmCampaign = append(e.UtmCampaign, n)
		}
		if n.UtmContent != "" {
			e.UtmContent = append(e.UtmContent, n)
		}
		if n.UtmTerm != "" {
			e.UtmTerm = append(e.UtmTerm, n)
		}
		if n.ScreenSize != "" {
			e.ScreenSize = append(e.ScreenSize, n)
		}
		if n.Browser != "" {
			e.Browser = append(e.Browser, n)
		}
		if n.BrowserVersion != "" {
			e.BrowserVersion = append(e.BrowserVersion, n)
		}
		if n.OperatingSystem != "" {
			e.OperatingSystem = append(e.OperatingSystem, n)
		}
		if n.OperatingSystemVersion != "" {
			e.OperatingSystemVersion = append(e.OperatingSystemVersion, n)
		}
		if n.CountryCode != "" {
			e.CountryCode = append(e.CountryCode, n)
		}
		if n.Region != "" {
			e.Region = append(e.Region, n)
		}
		if n.CityGeoNameId != "" {
			e.CityGeoNameId = append(e.CityGeoNameId, n)
		}
	}
}

func (e *EntrySegment) Reuse() {
	e.Name = e.Name[:0]
	e.Pathname = e.Pathname[:0]
	e.EntryPage = e.EntryPage[:0]
	e.ExitPage = e.ExitPage[:0]
	e.Referrer = e.Referrer[:0]
	e.ReferrerSource = e.ReferrerSource[:0]
	e.UtmMedium = e.UtmMedium[:0]
	e.UtmSource = e.UtmSource[:0]
	e.UtmCampaign = e.UtmCampaign[:0]
	e.UtmContent = e.UtmContent[:0]
	e.UtmTerm = e.UtmTerm[:0]
	e.ScreenSize = e.ScreenSize[:0]
	e.Browser = e.Browser[:0]
	e.BrowserVersion = e.BrowserVersion[:0]
	e.OperatingSystem = e.OperatingSystem[:0]
	e.OperatingSystemVersion = e.OperatingSystemVersion[:0]
	e.CountryCode = e.CountryCode[:0]
	e.Region = e.Region[:0]
	e.CityGeoNameId = e.CityGeoNameId[:0]
}

func (e *EntrySegment) Aggr(u, s *roaring64.Bitmap, o *Aggr) {
	aggr(e.Name, u, s, &o.Name,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.Name, e2.Name)
		},
		func(e *Entry) string {
			return e.Name
		},
	)
	aggr(e.Pathname, u, s, &o.Pathname,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.Pathname, e2.Pathname)
		},
		func(e *Entry) string {
			return e.Pathname
		},
	)
	aggr(e.EntryPage, u, s, &o.EntryPage,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.EntryPage, e2.EntryPage)
		},
		func(e *Entry) string {
			return e.EntryPage
		},
	)
	aggr(e.ExitPage, u, s, &o.ExitPage,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.ExitPage, e2.ExitPage)
		},
		func(e *Entry) string {
			return e.ExitPage
		},
	)
	aggr(e.Referrer, u, s, &o.Referrer,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.Referrer, e2.Referrer)
		},
		func(e *Entry) string {
			return e.Referrer
		},
	)
	aggr(e.ReferrerSource, u, s, &o.ReferrerSource,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.ReferrerSource, e2.ReferrerSource)
		},
		func(e *Entry) string {
			return e.ReferrerSource
		},
	)
	aggr(e.UtmMedium, u, s, &o.UtmMedium,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.UtmMedium, e2.UtmMedium)
		},
		func(e *Entry) string {
			return e.UtmMedium
		},
	)
	aggr(e.UtmSource, u, s, &o.UtmSource,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.UtmSource, e2.UtmSource)
		},
		func(e *Entry) string {
			return e.UtmSource
		},
	)
	aggr(e.UtmCampaign, u, s, &o.UtmCampaign,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.UtmCampaign, e2.UtmCampaign)
		},
		func(e *Entry) string {
			return e.UtmCampaign
		},
	)
	aggr(e.UtmContent, u, s, &o.UtmContent,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.UtmContent, e2.UtmContent)
		},
		func(e *Entry) string {
			return e.UtmContent
		},
	)
	aggr(e.UtmTerm, u, s, &o.UtmTerm,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.UtmTerm, e2.UtmTerm)
		},
		func(e *Entry) string {
			return e.UtmTerm
		},
	)
	aggr(e.ScreenSize, u, s, &o.ScreenSize,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.ScreenSize, e2.ScreenSize)
		},
		func(e *Entry) string {
			return e.ScreenSize
		},
	)
	aggr(e.Browser, u, s, &o.Browser,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.Browser, e2.Browser)
		},
		func(e *Entry) string {
			return e.Browser
		},
	)
	aggr(e.BrowserVersion, u, s, &o.BrowserVersion,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.BrowserVersion, e2.BrowserVersion)
		},
		func(e *Entry) string {
			return e.BrowserVersion
		},
	)
	aggr(e.OperatingSystem, u, s, &o.OperatingSystem,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.OperatingSystem, e2.OperatingSystem)
		},
		func(e *Entry) string {
			return e.OperatingSystem
		},
	)
	aggr(e.OperatingSystemVersion, u, s, &o.OperatingSystemVersion,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.OperatingSystemVersion, e2.OperatingSystemVersion)
		},
		func(e *Entry) string {
			return e.OperatingSystemVersion
		},
	)
	aggr(e.CountryCode, u, s, &o.CountryCode,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.CountryCode, e2.CountryCode)
		},
		func(e *Entry) string {
			return e.CountryCode
		},
	)
	aggr(e.Region, u, s, &o.Region,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.Region, e2.Region)
		},
		func(e *Entry) string {
			return e.Region
		},
	)
	aggr(e.CityGeoNameId, u, s, &o.CityGeoNameId,
		func(e1, e2 *Entry) int {
			return strings.Compare(e1.CityGeoNameId, e2.CityGeoNameId)
		},
		func(e *Entry) string {
			return e.CityGeoNameId
		},
	)

}

func aggr(ls []*Entry, u,
	s *roaring64.Bitmap,
	o **Aggr_Segment,
	cmp func(*Entry, *Entry) int,
	key func(*Entry) string,
) {
	if len(ls) > 0 {
		*o = &Aggr_Segment{
			Aggregates: make(map[string]*Aggr_Total),
		}
		w := *o
		// sort
		sort.Slice(ls, func(i, j int) bool {
			return cmp(ls[i], ls[j]) == -1
		})
		var pos int
		for i := range ls {
			if i > 0 {
				if cmp(ls[i], ls[i-1]) != 0 {
					w.Aggregates[key(ls[i-1])] = EntryList(ls[pos:i]).Aggr(u, s)
					pos = i
				}
			}
		}
		if pos < len(ls)-1 {
			w.Aggregates[key(ls[pos])] = EntryList(ls[pos:]).Aggr(u, s)
		}
	}
}

func updateRoot(txn *badger.Txn, ts time.Time, id *ID, a *Aggr_Total) error {
	var key []byte
	for i := TABLE_RAW; i <= TABLE_YEAR; i += 1 {
		switch i {
		case TABLE_RAW:
			key = id.Raw().SetTable(byte(i)).SetMeta(0).Final()
		case TABLE_HOUR:
			key = id.Hour(ts).SetTable(byte(i)).SetMeta(0).Final()
		case TABLE_DAY:
			key = id.Day(ts).SetTable(byte(i)).SetMeta(0).Final()
		case TABLE_MONTH:
			key = id.Month(ts).SetTable(byte(i)).SetMeta(0).Final()
		case TABLE_YEAR:
			key = id.Year(ts).SetTable(byte(i)).SetMeta(0).Final()
		}
		err := updateAggregate(txn, key, a)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateMeta(txn *badger.Txn, ts time.Time, id *ID, meta PROPS, a *Aggr_Segment) error {
	if a == nil {
		return nil
	}
	var key []byte
	for i := TABLE_HOUR; i <= TABLE_YEAR; i += 1 {
		switch i {
		case TABLE_HOUR:
			key = id.Hour(ts).SetTable(byte(i)).SetMeta(byte(meta)).Final()
		case TABLE_DAY:
			key = id.Day(ts).SetTable(byte(i)).SetMeta(byte(meta)).Final()
		case TABLE_MONTH:
			key = id.Month(ts).SetTable(byte(i)).SetMeta(byte(meta)).Final()
		case TABLE_YEAR:
			key = id.Year(ts).SetTable(byte(i)).SetMeta(byte(meta)).Final()
		}
		err := updateSegment(txn, key, a)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateAggregate(txn *badger.Txn, key []byte, a *Aggr_Total) error {
	x, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			// new entry we store a
			b, err := proto.Marshal(a)
			if err != nil {
				return fmt.Errorf("failed to marshal aggr total %v", err)
			}
			return txn.Set(key, b)
		}
		return err
	}
	var o Aggr_Total
	err = x.Value(func(val []byte) error {
		return proto.Unmarshal(val, &o)
	})
	if err != nil {
		return fmt.Errorf("failed to read  aggr total %v", err)
	}
	o.Add(a)
	b, err := proto.Marshal(&o)
	if err != nil {
		return fmt.Errorf("failed to marshal aggr total %v", err)
	}
	return txn.Set(key, b)
}

func updateSegment(txn *badger.Txn, key []byte, a *Aggr_Segment) error {
	x, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			// new entry we store a
			b, err := proto.Marshal(a)
			if err != nil {
				return fmt.Errorf("failed to marshal aggr total %v", err)
			}
			return txn.Set(key, b)
		}
		return err
	}
	var o Aggr_Segment
	err = x.Value(func(val []byte) error {
		return proto.Unmarshal(val, &o)
	})
	if err != nil {
		return fmt.Errorf("failed to read  aggr total %v", err)
	}
	o.Add(a)
	b, err := proto.Marshal(&o)
	if err != nil {
		return fmt.Errorf("failed to marshal aggr total %v", err)
	}
	return txn.Set(key, b)
}

func (a *Aggr_Total) Add(o *Aggr_Total) {
	a.Visitors += o.Visitors
	a.Visits += o.Visits
	a.Events += o.Events
}

func (a *Aggr_Segment) Add(o *Aggr_Segment) {
	for k, v := range o.Aggregates {
		if x, ok := a.Aggregates[k]; ok {
			x.Add(v)
		} else {
			a.Aggregates[k] = v
		}
	}
}
