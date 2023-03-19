package timeseries

import (
	"context"
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

type entryBuf struct {
	entries EntryList
}

func (e *entryBuf) ensure(capacity int) []*Entry {
	if cap(e.entries) < capacity {
		e.entries = append([]*Entry(nil), make([]*Entry, capacity)...)
	}
	return e.entries[:capacity]
}

func (e *entryBuf) Release() {
	e.entries = e.entries[:0]
}

var entryBufPool = &sync.Pool{
	New: func() any {
		return &entryBuf{
			entries: make([]*Entry, 1024),
		}
	},
}

type Group struct {
	props [PROPS_CITY]*mapEntry
}

var groupPool = &sync.Pool{
	New: func() any {
		return newGroup()
	},
}

func newGroup() *Group {
	var g Group
	for i := range g.props {
		g.props[i] = &mapEntry{
			m: make(map[string]*entryBuf),
		}
	}
	return &g
}

func (g *Group) save(p PROPS, key string, e *Entry) {
	g.props[p].save(key, e)
}

func (g *Group) Save(h int, ts time.Time, ms *MetricSaver) {
	for i, e := range g.props {
		ms.Save(h, ts, func(u, s *roaring64.Bitmap, hs *HourStats) {
			if hs.Properties[uint32(i)] == nil {
				hs.Properties[uint32(i)] = &PropAggregate{}
			}
			e.Save(u, s, hs.Properties[uint32(i)])
		})
	}
}

func (g *Group) Reset() {
	for _, e := range g.props {
		e.Release()
	}
}

func (g *Group) Release() {
	g.Reset()
	groupPool.Put(g)
}

type mapEntry struct {
	m map[string]*entryBuf
}

func (m *mapEntry) Release() {
	if m.m != nil {
		for k, v := range m.m {
			v.Release()
			delete(m.m, k)
		}
	}
}

func (m *mapEntry) Save(u, s *roaring64.Bitmap, e *PropAggregate) {
	if m.m != nil {
		for k, v := range m.m {
			if e.Aggregate == nil {
				e.Aggregate = make(map[string]*Aggregate)
			}
			e.Aggregate[k] = v.entries.Aggregate(u, s)
		}
	}
}

func (m *mapEntry) save(key string, e *Entry) {
	if key == "" {
		return
	}
	if b, ok := m.m[key]; ok {
		b.entries = append(b.entries, e)
		return
	}
	b := getEntryBuf()
	m.m[key] = b
	b.entries = append(b.entries, e)
}

func getEntryBuf() *entryBuf {
	e := entryBufPool.Get().(*entryBuf)
	e.entries = e.entries[:cap(e.entries)]
	return e
}

// Process and save data from b to m. Daily data is continuously merged until the next date
// Final daily data is packed into parquet file for permanent storage.
func (m *Mike) Save(ctx context.Context, b *Buffer, uid, sid uint64) {
	defer b.Release()
	ms := metricSaverPool.Get().(*MetricSaver)
	defer ms.Release()
	group := groupPool.Get().(*Group)

	ent := EntryList(b.entries)
	ent.Emit(func(i int, el EntryList) {
		// reset for group works differently. It only release aggregate buffers
		// accumulated in this callback.
		//
		// We avoid allocation property array(managed by group) for every iteration
		// we only pay the prices of deleting map entries for the aggregate properties
		// data that we have done processing
		//
		// TODO: never delete released buffers, just mark them as deleted after
		// reset.
		defer group.Reset()

		// By now we have el which is a list of all entries happened in i hour
		// We segment the props accordingly and compute aggregates That we save
		// on saver.
		for _, e := range el {
			group.save(PROPS_NAME, e.Name, e)
			group.save(PROPS_PAGE, e.Pathname, e)
			group.save(PROPS_ENTRY_PAGE, e.EntryPage, e)
			group.save(PROPS_EXIT_PAGE, e.ExitPage, e)
			group.save(PROPS_REFERRER, e.Referrer, e)
			group.save(PROPS_UTM_MEDIUM, e.UtmMedium, e)
			group.save(PROPS_UTM_SOURCE, e.UtmSource, e)
			group.save(PROPS_UTM_CAMPAIGN, e.UtmCampaign, e)
			group.save(PROPS_UTM_CONTENT, e.UtmContent, e)
			group.save(PROPS_UTM_TERM, e.UtmTerm, e)
			group.save(PROPS_UTM_DEVICE, e.ScreenSize, e)
			group.save(PROPS_UTM_BROWSER, e.Browser, e)
			group.save(PROPS_BROWSER_VERSION, e.BrowserVersion, e)
			group.save(PROPS_OS, e.OperatingSystem, e)
			group.save(PROPS_OS_VERSION, e.OperatingSystemVersion, e)
			group.save(PROPS_COUNTRY, e.CountryCode, e)
			group.save(PROPS_REGION, e.Region, e)
			group.save(PROPS_CITY, e.CityGeoNameId, e)
		}
		group.Save(i, time.Unix(el[0].Timestamp, 0), ms)
		ms.UpdateHourTotals(i, el)
	})

	id := NewID()
	defer id.Release()
	id.SetSiteID(sid)
	id.SetUserID(uid)

	for i := range ms.prop {
		h := &ms.prop[i]
		if len(h.Properties) > 0 {
			id.SetDate(h.Timestamp.AsTime())
			id.SetEntropy()
			err := m.db.Update(func(txn *badger.Txn) error {
				b, err := proto.Marshal(h)
				if err != nil {
					return err
				}
				// This will later be discarded after we have aggregated all hourly
				// stats. Keep them not more than 3 hours.
				//
				// NOTE: The key is by date. We use  meta to mark this as a hourly
				// record.
				e := badger.NewEntry(id[:], b).
					WithMeta(byte(Hour)).
					WithTTL(3 * time.Hour)
				return txn.SetEntry(e)
			})
			if err != nil {
				log.Get(ctx).Err(err).Msg("failed to save hourly stats ")
			}
		}
	}
}
