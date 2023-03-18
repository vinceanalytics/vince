package timeseries

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/dgraph-io/badger/v3"
	"github.com/segmentio/parquet-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
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
			entries: make([]*Entry, 4098),
		}
	},
}

type Group struct {
	props [PROPS_CITY]*mapEntry
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
	b := entryBufPool.Get().(*entryBuf)
	m.m[key] = b
	b.entries = append(b.entries, e)
}

// useful for reading. Returns entryBuf with entries spanning full capacity.
func getEntryBuf() *entryBuf {
	e := entryBufPool.Get().(*entryBuf)
	e.entries = e.entries[:cap(e.entries)]
	return e
}

// Aggregate computes aggregate stats segmented by the hour of the day.
func (m *Mike) Aggregate(ctx context.Context, id *ID, r io.ReaderAt, size int64) (day *DayStats, err error) {
	f, err := parquet.OpenFile(r, size)
	if err != nil {
		return nil, err
	}
	if f.NumRows() == 0 {
		// Unlikely but ok
		id.Release()
		return nil, nil
	}
	day = &DayStats{
		Aggregate: &Aggregate{},
	}
	saver := metricSaverPool.Get().(*MetricSaver)
	e := getEntryBuf()
	group := NewGroup()
	defer func() {
		saver.Release()
		e.Release()
		group.Release()
	}()

	for _, rg := range f.RowGroups() {
		err := m.writeRowGroup(id, rg, saver, e, group)
		if err != nil {
			return nil, err
		}
	}
	var d time.Duration
	var visit float64
	for i := range saver.prop {
		h := &saver.prop[i]
		day.Aggregate.Visitors += h.Aggregate.Visitors
		day.Aggregate.Visitors += h.Aggregate.Visits
		d += h.Aggregate.VisitDuration.AsDuration()
		visit += h.Aggregate.ViewsPerVisit
		day.Stats = append(day.Stats, proto.Clone(h).(*HourStats))
	}
	day.Aggregate.ViewsPerVisit = visit / 12
	day.Aggregate.VisitDuration = durationpb.New(d / 12)
	return
}

func (m *Mike) writeRowGroup(
	id *ID,
	g parquet.RowGroup,
	saver *MetricSaver,
	e *entryBuf,
	group *Group,
) error {
	r := parquet.NewGenericRowGroupReader[*Entry](g)
	values := e.ensure(int(r.NumRows()))
	n, err := r.Read(values)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	ent := EntryList(values[:n])

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
		group.Save(i, time.Unix(el[0].Timestamp, 0), saver)
		saver.UpdateHourTotals(i, el)
	})
	return nil
}
