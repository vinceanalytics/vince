package timeseries

import (
	"context"
	"errors"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/dgraph-io/badger/v3"
	"github.com/segmentio/parquet-go"
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

func (e *entryBuf) emit(f func(int, EntryList)) {
	if len(e.entries) == 0 {
		return
	}
	var start int
	for i := range e.entries {
		if i > 0 && e.entries[i].Timestamp.Hour() != e.entries[i-1].Timestamp.Hour() {
			f(e.entries[start].Timestamp.Hour(), e.entries[start:i])
			start = i
		}
	}
	if start < len(e.entries) {
		f(e.entries[start].Timestamp.Hour(), e.entries[start:])
	}
}

func (e *entryBuf) Release() {
	e.entries = e.entries[:0]
}

var entryBufPool = &sync.Pool{
	New: func() any {
		return &entryBuf{
			entries: make([]*Entry, 4098),
		}
	},
}

type Group struct {
	props [CITY + 1]*mapEntry
}

func NewGroup() *Group {
	var g Group
	for i := range g.props {
		g.props[i] = &mapEntry{
			m: make(map[string]*entryBuf),
		}
	}
	return &g
}

func (g *Group) save(p PROPERTY, key string, e *Entry) {
	g.props[p].save(key, e)
}

func (g *Group) saveInt(p PROPERTY, key uint32, e *Entry) {
	g.props[p].saveInt(key, e)
}

func (g *Group) Save(h int, ts time.Time, ms *MetricSaver) {
	for i, e := range g.props {
		ms.Save(h, ts, func(b *roaring64.Bitmap, hs *HourStats) {
			p := hs.Properties
			var agg *PropAggregate
			switch PROPERTY(i) {
			case NAME:
				if p.Name == nil {
					p.Name = &PropAggregate{}
				}
				agg = p.Name
			case PAGE:
				if p.Page == nil {
					p.Page = &PropAggregate{}
				}
				agg = p.Page
			case ENTRY_PAGE:
				if p.EntryPage == nil {
					p.EntryPage = &PropAggregate{}
				}
				agg = p.EntryPage
			case EXIT_PAGE:
				if p.ExitPage == nil {
					p.ExitPage = &PropAggregate{}
				}
				agg = p.ExitPage
			case REFERRER:
				if p.Referrer == nil {
					p.Referrer = &PropAggregate{}
				}
				agg = p.Referrer
			case UTM_MEDIUM:
				if p.UtmMedium == nil {
					p.UtmMedium = &PropAggregate{}
				}
				agg = p.UtmMedium
			case UTM_SOURCE:
				if p.UtmSource == nil {
					p.UtmSource = &PropAggregate{}
				}
				agg = p.UtmSource
			case UTM_CAMPAIGN:
				if p.UtmCampaign == nil {
					p.UtmCampaign = &PropAggregate{}
				}
				agg = p.UtmCampaign
			case UTM_CONTENT:
				if p.UtmContent == nil {
					p.UtmContent = &PropAggregate{}
				}
				agg = p.UtmContent
			case UTM_TERM:
				if p.UtmTerm == nil {
					p.UtmTerm = &PropAggregate{}
				}
				agg = p.UtmTerm
			case UTM_DEVICE:
				if p.UtmDevice == nil {
					p.UtmDevice = &PropAggregate{}
				}
				agg = p.UtmDevice
			case BROWSER:
				if p.Browser == nil {
					p.Browser = &PropAggregate{}
				}
				agg = p.Browser
			case BROWSER_VERSION:
				if p.BrowserVersion == nil {
					p.BrowserVersion = &PropAggregate{}
				}
				agg = p.BrowserVersion
			case OS:
				if p.Os == nil {
					p.Os = &PropAggregate{}
				}
				agg = p.Os
			case OS_VERSION:
				if p.OsVersion == nil {
					p.OsVersion = &PropAggregate{}
				}
				agg = p.OsVersion
			case COUNTRY:
				if p.Country == nil {
					p.Country = &PropAggregate{}
				}
				agg = p.Country
			case REGION:
				if p.Region == nil {
					p.Region = &PropAggregate{}
				}
				agg = p.Region
			case CITY:
				if p.City == nil {
					p.City = &PropAggregate{}
				}
				agg = p.City
			}
			e.Save(b, agg)
		})
	}
}

func (g *Group) Release() {
	for _, e := range g.props {
		e.Release()
	}
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

func (m *mapEntry) Save(r *roaring64.Bitmap, e *PropAggregate) {
	if m.m != nil {
		for k, v := range m.m {
			if e.Aggregate == nil {
				e.Aggregate = make(map[string]*Aggregate)
			}
			e.Aggregate[k] = &Aggregate{
				Visitors: v.entries.Visitors(r),
				Visits:   v.entries.Visits(r),
			}
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
	b := entryBufPool.Get().(*entryBuf)
	m.m[key] = b
	b.entries = append(b.entries, e)
}

func (m *mapEntry) saveInt(key uint32, e *Entry) {
	if key == 0 {
		return
	}
	k := strconv.FormatUint(uint64(key), 10)
	if b, ok := m.m[k]; ok {
		b.entries = append(b.entries, e)
		return
	}
	b := entryBufPool.Get().(*entryBuf)
	m.m[k] = b
	b.entries = append(b.entries, e)
}

// useful for reading. Returns entryBuf with entries spanning full capacity.
func getEntryBuf() *entryBuf {
	e := entryBufPool.Get().(*entryBuf)
	e.entries = e.entries[:cap(e.entries)]
	return e
}

// Write process file and stores it permanently. file is assumed not to be owned
// by m. This is supposed to happen in the background.
//
// RowGroups are processed sequentially. Concurrency is per file/id
func (m *Mike) Write(ctx context.Context, id *ID, r io.ReaderAt, size int64) error {
	f, err := parquet.OpenFile(r, size)
	if err != nil {
		return err
	}
	if f.NumRows() == 0 {
		// Unlikely but ok
		id.Release()
		return nil
	}
	saver := metricSaverPool.Get().(*MetricSaver)
	e := getEntryBuf()
	group := NewGroup()
	defer func() {
		saver.Release()
		e.Release()
		group.Release()
	}()

	var totalVisitors, totalVisis uint64
	for _, rg := range f.RowGroups() {
		visitors, visits, err := m.writeRowGroup(id, rg, saver, e, group)
		if err != nil {
			return err
		}
		totalVisitors += visitors
		totalVisis += visits
	}
	return nil
}

func (m *Mike) writeRowGroup(
	id *ID,
	g parquet.RowGroup,
	saver *MetricSaver,
	e *entryBuf,
	group *Group,
) (visitors, visits uint64, err error) {

	r := parquet.NewGenericRowGroupReader[*Entry](g)
	values := e.ensure(int(r.NumRows()))
	n, err := r.Read(values)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return 0, 0, err
		}
	}

	ent := EntryList(values[:n])

	// calculate totals and ensure hash for sessions is computed here. Next calls
	// to visit will use computed hash.
	visitors = ent.Visitors(saver.bloom)
	visits = ent.Visits(saver.bloom)

	ent.Emit(func(i int, el EntryList) {
		// release for group works differently. It only release aggregate buffers
		// accumulated in this callback.
		//
		// We avoid allocation property array(managed by group) for every iteration
		// we only pay the prices of deleting map entries for the aggregate properties
		// data.
		defer group.Release()

		// By now we have el which is a list of all entries happened in i hour
		// We segment the props accordingly and compute aggregates That we save
		// on saver.
		for _, e := range el[:n] {
			group.save(NAME, e.Name, e)
			group.save(PAGE, e.Pathname, e)
			group.save(ENTRY_PAGE, e.EntryPage, e)
			group.save(EXIT_PAGE, e.ExitPage, e)
			group.save(REFERRER, e.Referrer, e)
			group.save(UTM_MEDIUM, e.UtmMedium, e)
			group.save(UTM_SOURCE, e.UtmSource, e)
			group.save(UTM_CAMPAIGN, e.UtmCampaign, e)
			group.save(UTM_CONTENT, e.UtmContent, e)
			group.save(UTM_TERM, e.UtmTerm, e)
			group.save(UTM_DEVICE, e.ScreenSize, e)
			group.save(BROWSER, e.Browser, e)
			group.save(BROWSER_VERSION, e.BrowserVersion, e)
			group.save(OS, e.OperatingSystem, e)
			group.save(OS_VERSION, e.OperatingSystemVersion, e)
			group.save(COUNTRY, e.CountryCode, e)
			group.save(REGION, e.Region, e)
			group.saveInt(CITY, e.CityGeoNameID, e)
		}
		group.Save(i, el[0].Timestamp, saver)
		saver.UpdateHourTotals(i, el)
	})
	return
}
