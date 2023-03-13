package timeseries

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/dgraph-io/badger/v3"
	"github.com/segmentio/parquet-go"
	"golang.org/x/sync/errgroup"
)

// Mike is the permanent storage for the events data. Data stored here is aggregated
// and broken down. All data is still stored in parquet format. This  only supports
// reads and writes, nothing is ever deleted from this storage.
type Mike struct {
	db *badger.DB
}

type entryBuf struct {
	entries []*Entry
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

type mapEntry struct {
	prop PROPERTY
	m    map[string]*entryBuf
	m2   map[uint32]*entryBuf
}

func (m *mapEntry) Release() {
	if m.m != nil {
		for _, v := range m.m {
			v.Release()
		}
	}
	if m.m2 != nil {
		for _, v := range m.m2 {
			v.Release()
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
	if b, ok := m.m2[key]; ok {
		b.entries = append(b.entries, e)
		return
	}
	b := entryBufPool.Get().(*entryBuf)
	m.m2[key] = b
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
// RowGroups are processed concurrently.
func (m *Mike) Write(ctx context.Context, id *ID, r io.ReaderAt, size int64) error {
	f, err := parquet.OpenFile(r, size)
	if err != nil {
		return err
	}
	if f.NumRows() == 0 {
		// Unlikely but ok
		return nil
	}
	g, ctx := errgroup.WithContext(ctx)
	for _, rg := range f.RowGroups() {
		g.Go(m.writeRowGroup(ctx, id.Clone(), rg))
	}
	return g.Wait()
}

func (m *Mike) writeRowGroup(ctx context.Context, id *ID, g parquet.RowGroup) func() error {
	return func() error {
		e := getEntryBuf()
		var resources releaseList
		resources = append(resources, e)
		resources = append(resources, id)
		defer resources.Release()

		group := make([]*mapEntry, CITY+1)

		for i := 0; i < len(group); i++ {
			if i == int(CITY) {
				group[i] = &mapEntry{prop: PROPERTY(i), m2: make(map[uint32]*entryBuf)}
			} else {
				group[i] = &mapEntry{prop: PROPERTY(i), m: make(map[string]*entryBuf)}
			}
			resources = append(resources, group[i])
		}
		values := e.entries
		r := parquet.NewGenericRowGroupReader[*Entry](g)
		for {
			n, err := r.Read(values)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					return err
				}
			}
			if errors.Is(err, io.EOF) {
				break
			}
			for _, e := range values[:n] {
				group[NAME].save(e.Name, e)
				group[PAGE].save(e.Pathname, e)
				group[ENTRY_PAGE].save(e.EntryPage, e)
				group[EXIT_PAGE].save(e.ExitPage, e)
				group[REFERRER].save(e.Referrer, e)
				group[UTM_MEDIUM].save(e.UtmMedium, e)
				group[UTM_SOURCE].save(e.UtmSource, e)
				group[UTM_CAMPAIGN].save(e.UtmCampaign, e)
				group[UTM_CONTENT].save(e.UtmContent, e)
				group[UTM_TERM].save(e.UtmTerm, e)
				group[UTM_DEVICE].save(e.ScreenSize, e)
				group[BROWSER].save(e.Browser, e)
				group[BROWSER_VERSION].save(e.BrowserVersion, e)
				group[OS].save(e.OperatingSystem, e)
				group[OS_VERSION].save(e.OperatingSystemVersion, e)
				group[COUNTRY].save(e.CountryCode, e)
				group[REGION].save(e.Region, e)
				group[CITY].saveInt(e.CityGeoNameID, e)
			}
		}
		return nil
	}

}
