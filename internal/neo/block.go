package neo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/parquet-go/parquet-go"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/blocks"
	"github.com/vinceanalytics/vince/pkg/entry"
	"github.com/vinceanalytics/vince/pkg/log"
	"google.golang.org/protobuf/proto"
)

const (
	FilterBitsPerValue = 10
	BlockFile          = "BLOCK"
)

var metaPath = make([]byte, 1)

type ActiveBlock struct {
	mu       sync.Mutex
	bloom    metaBloom
	Min, Max time.Time
	dir      string
	f        *os.File
	b        bytes.Buffer
	db       *badger.DB
	// rows are already buffered with w. It is wise to send entries to w as they
	// arrive. tmp allows us to avoid creating a slice with one entry on every
	// WriteRow call
	tmp [1]*entry.Entry
	w   *parquet.SortingWriter[*entry.Entry]
	n   int
}

func NewBlock(dir string, db *badger.DB) (*ActiveBlock, error) {
	a := &ActiveBlock{
		bloom: metaBloom{hash: xxhash.New()},
		db:    db,
		dir:   dir,
	}
	return a, a.open()
}

func (a *ActiveBlock) Save() error {
	a.mu.Lock()
	err := a.open()
	a.mu.Unlock()
	return err
}

func (a *ActiveBlock) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return errors.Join(a.open(), a.f.Close())
}

func (a *ActiveBlock) open() error {
	if a.f != nil {
		if a.n != 0 {
			a.w.Close()
			_, err := a.f.Seek(0, io.SeekStart)
			if err != nil {
				return err
			}
			a.b.Reset()
			io.Copy(&a.b, a.f)
			err = a.save()
			if err != nil {
				return err
			}
			a.bloom.reset()
			a.n = 0
		}
		a.f.Close()
	}
	var err error
	a.f, err = os.Create(filepath.Join(a.dir, BlockFile))
	if err != nil {
		return err
	}
	if a.w == nil {
		a.w = Writer[*entry.Entry](a.f)
	}
	a.w.Reset(a.f)
	return nil
}

func (a *ActiveBlock) WriteEntry(e *entry.Entry) {
	a.mu.Lock()
	if a.Min.IsZero() {
		a.Min = e.Timestamp
	}
	a.Max = e.Timestamp
	a.tmp[0] = e.Clone()
	n, err := a.w.Write(a.tmp[:])
	if err != nil {
		log.Get().Fatal().Err(err).
			Msg("failed to save entries")
	}
	a.n += n
	a.bloom.set(e)
	a.mu.Unlock()
}

func (a *ActiveBlock) save() error {
	return a.db.Update(func(txn *badger.Txn) error {
		meta := &blocks.Metadata{}
		x, err := txn.Get([]byte(metaPath))
		if err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
		} else {
			err = x.Value(func(val []byte) error {
				return proto.Unmarshal(val, meta)
			})
			if err != nil {
				return err
			}
		}
		id := time.Now().UTC().UnixMilli()
		blockPath := make([]byte, 8)
		binary.BigEndian.PutUint64(blockPath, uint64(id))
		meta.Blocks = append(meta.Blocks, &blocks.Block{
			Id:    id,
			Min:   a.Min.UnixMilli(),
			Max:   a.Max.UnixMilli(),
			Size:  int64(a.b.Len()),
			Bloom: a.bloom.bloom(),
		})
		mb, err := proto.Marshal(meta)
		if err != nil {
			return err
		}
		return errors.Join(
			txn.Set(blockPath, a.b.Bytes()),
			txn.Set(metaPath, mb),
		)
	})
}

// Writer returns a parquet.SortingWriter for T that sorts timestamp field in
// ascending order.
func Writer[T any](w io.Writer, o ...parquet.WriterOption) *parquet.SortingWriter[T] {
	var t T
	scheme := parquet.SchemaOf(t)
	var bloom []parquet.BloomFilterColumn
	for _, col := range scheme.Columns() {
		l, _ := scheme.Lookup(col...)
		if l.Node.Type().Kind() == parquet.ByteArray {
			bloom = append(bloom, parquet.SplitBlockFilter(FilterBitsPerValue, col...))
		}
	}
	return parquet.NewSortingWriter[T](w, 4<<10, append(o,
		parquet.BloomFilters(bloom...),
		parquet.SortingWriterConfig(
			parquet.SortingColumns(
				parquet.Ascending("timestamp"),
			),
		),
	)...)
}

type metaBloom struct {
	hash           *xxhash.Digest
	Browser        *roaring64.Bitmap
	BrowserVersion *roaring64.Bitmap
	City           *roaring64.Bitmap
	Country        *roaring64.Bitmap
	Domain         *roaring64.Bitmap
	EntryPage      *roaring64.Bitmap
	ExitPage       *roaring64.Bitmap
	Host           *roaring64.Bitmap
	Name           *roaring64.Bitmap
	Os             *roaring64.Bitmap
	OsVersion      *roaring64.Bitmap
	Path           *roaring64.Bitmap
	Referrer       *roaring64.Bitmap
	ReferrerSource *roaring64.Bitmap
	Region         *roaring64.Bitmap
	Screen         *roaring64.Bitmap
	UtmCampaign    *roaring64.Bitmap
	UtmContent     *roaring64.Bitmap
	UtmMedium      *roaring64.Bitmap
	UtmSource      *roaring64.Bitmap
	UtmTerm        *roaring64.Bitmap
	UtmValue       *roaring64.Bitmap
}

func (m *metaBloom) reset() {
	// we only reuse hash digest
	h := m.hash
	h.Reset()
	pubBitmaps(
		m.Browser,
		m.BrowserVersion,
		m.City,
		m.Country,
		m.Domain,
		m.EntryPage,
		m.ExitPage,
		m.Host,
		m.Name,
		m.Os,
		m.OsVersion,
		m.Path,
		m.Referrer,
		m.ReferrerSource,
		m.Region,
		m.Screen,
		m.UtmCampaign,
		m.UtmContent,
		m.UtmMedium,
		m.UtmSource,
		m.UtmTerm,
		m.UtmValue,
	)
	*m = metaBloom{}
	m.hash = h
}

func (m *metaBloom) sum(s string) uint64 {
	m.hash.Reset()
	m.hash.WriteString(s)
	return m.hash.Sum64()
}

func (m *metaBloom) set(e *entry.Entry) {
	if e.Browser != "" {
		if m.Browser == nil {
			m.Browser = newBitmap()
		}
		m.Browser.Add(m.sum(e.Browser))
	}
	if e.BrowserVersion != "" {
		if m.BrowserVersion == nil {
			m.BrowserVersion = newBitmap()
		}
		m.BrowserVersion.Add(m.sum(e.BrowserVersion))
	}
	if e.City != "" {
		if m.City == nil {
			m.City = newBitmap()
		}
		m.City.Add(m.sum(e.City))
	}
	if e.Country != "" {
		if m.Country == nil {
			m.Country = newBitmap()
		}
		m.Country.Add(m.sum(e.Country))
	}
	if e.Domain != "" {
		if m.Domain == nil {
			m.Domain = newBitmap()
		}
		m.Domain.Add(m.sum(e.Domain))
	}
	if e.EntryPage != "" {
		if m.EntryPage == nil {
			m.EntryPage = newBitmap()
		}
		m.EntryPage.Add(m.sum(e.EntryPage))
	}
	if e.ExitPage != "" {
		if m.ExitPage == nil {
			m.ExitPage = newBitmap()
		}
		m.ExitPage.Add(m.sum(e.ExitPage))
	}
	if e.Host != "" {
		if m.Host == nil {
			m.Host = newBitmap()
		}
		m.Host.Add(m.sum(e.Host))
	}
	if e.Name != "" {
		if m.Name == nil {
			m.Name = newBitmap()
		}
		m.Name.Add(m.sum(e.Name))
	}
	if e.Os != "" {
		if m.Os == nil {
			m.Os = newBitmap()
		}
		m.Os.Add(m.sum(e.Os))
	}
	if e.OsVersion != "" {
		if m.OsVersion == nil {
			m.OsVersion = newBitmap()
		}
		m.OsVersion.Add(m.sum(e.OsVersion))
	}
	if e.Path != "" {
		if m.Path == nil {
			m.Path = newBitmap()
		}
		m.Path.Add(m.sum(e.Path))
	}
	if e.Referrer != "" {
		if m.Referrer == nil {
			m.Referrer = newBitmap()
		}
		m.Path.Add(m.sum(e.Referrer))
	}
	if e.ReferrerSource != "" {
		if m.ReferrerSource == nil {
			m.ReferrerSource = newBitmap()
		}
		m.ReferrerSource.Add(m.sum(e.ReferrerSource))
	}
	if e.Region != "" {
		if m.Region == nil {
			m.Region = newBitmap()
		}
		m.ReferrerSource.Add(m.sum(e.ReferrerSource))
	}
	if e.Screen != "" {
		if m.Screen == nil {
			m.Screen = newBitmap()
		}
		m.Screen.Add(m.sum(e.Screen))
	}
	if e.UtmCampaign != "" {
		if m.UtmCampaign == nil {
			m.UtmCampaign = newBitmap()
		}
		m.UtmCampaign.Add(m.sum(e.UtmCampaign))
	}
	if e.UtmContent != "" {
		if m.UtmContent == nil {
			m.UtmContent = newBitmap()
		}
		m.UtmContent.Add(m.sum(e.UtmContent))
	}
	if e.UtmMedium != "" {
		if m.UtmMedium == nil {
			m.UtmMedium = newBitmap()
		}
		m.UtmMedium.Add(m.sum(e.UtmMedium))
	}
	if e.UtmMedium != "" {
		if m.UtmMedium == nil {
			m.UtmMedium = newBitmap()
		}
		m.UtmMedium.Add(m.sum(e.UtmMedium))
	}
	if e.UtmSource != "" {
		if m.UtmSource == nil {
			m.UtmSource = newBitmap()
		}
		m.UtmSource.Add(m.sum(e.UtmSource))
	}
	if e.UtmTerm != "" {
		if m.UtmTerm == nil {
			m.UtmTerm = newBitmap()
		}
		m.UtmTerm.Add(m.sum(e.UtmTerm))
	}
	if e.UtmTerm != "" {
		if m.UtmTerm == nil {
			m.UtmTerm = newBitmap()
		}
		m.UtmTerm.Add(m.sum(e.UtmTerm))
	}
}

func (m *metaBloom) bloom() (b *blocks.Bloom) {
	b = &blocks.Bloom{}
	if m.Browser != nil {
		b.Browser = must.Must(m.Browser.MarshalBinary())
	}
	if m.BrowserVersion != nil {
		b.BrowserVersion = must.Must(m.BrowserVersion.MarshalBinary())
	}
	if m.City != nil {
		b.City = must.Must(m.City.MarshalBinary())
	}
	if m.Country != nil {
		b.Country = must.Must(m.Country.MarshalBinary())
	}
	if m.Domain != nil {
		b.Domain = must.Must(m.Domain.MarshalBinary())
	}
	if m.EntryPage != nil {
		b.EntryPage = must.Must(m.EntryPage.MarshalBinary())
	}
	if m.ExitPage != nil {
		b.ExitPage = must.Must(m.ExitPage.MarshalBinary())
	}
	if m.Host != nil {
		b.Host = must.Must(m.Host.MarshalBinary())
	}
	if m.Name != nil {
		b.Name = must.Must(m.Name.MarshalBinary())
	}
	if m.Os != nil {
		b.Os = must.Must(m.Os.MarshalBinary())
	}
	if m.OsVersion != nil {
		b.OsVersion = must.Must(m.OsVersion.MarshalBinary())
	}
	if m.Path != nil {
		b.Path = must.Must(m.Path.MarshalBinary())
	}
	if m.Referrer != nil {
		b.Referrer = must.Must(m.Referrer.MarshalBinary())
	}
	if m.ReferrerSource != nil {
		b.ReferrerSource = must.Must(m.ReferrerSource.MarshalBinary())
	}
	if m.Region != nil {
		b.Region = must.Must(m.Region.MarshalBinary())
	}
	if m.Screen != nil {
		b.Screen = must.Must(m.Screen.MarshalBinary())
	}
	if m.UtmCampaign != nil {
		b.UtmCampaign = must.Must(m.UtmCampaign.MarshalBinary())
	}
	if m.UtmContent != nil {
		b.UtmContent = must.Must(m.UtmContent.MarshalBinary())
	}
	if m.UtmMedium != nil {
		b.UtmMedium = must.Must(m.UtmMedium.MarshalBinary())
	}
	if m.UtmMedium != nil {
		b.UtmMedium = must.Must(m.UtmMedium.MarshalBinary())
	}
	if m.UtmSource != nil {
		b.UtmSource = must.Must(m.Browser.MarshalBinary())
	}
	if m.UtmTerm != nil {
		b.UtmTerm = must.Must(m.UtmTerm.MarshalBinary())
	}
	if m.UtmTerm != nil {
		b.UtmTerm = must.Must(m.UtmTerm.MarshalBinary())
	}
	return
}

var roaringPool = &sync.Pool{
	New: func() any {
		return roaring64.New()
	},
}

func newBitmap() *roaring64.Bitmap {
	return roaringPool.Get().(*roaring64.Bitmap)
}

func pubBitmaps(m ...*roaring64.Bitmap) {
	for _, r := range m {
		if r != nil {
			r.Clear()
			roaringPool.Put(r)
		}
	}
}
