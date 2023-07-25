package neo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v13/arrow"
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/parquet-go/parquet-go"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/blocks"
	"github.com/vinceanalytics/vince/pkg/entry"
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
	b        bytes.Buffer
	db       *badger.DB
	entries  *entry.MultiEntry
}

func NewBlock(dir string, db *badger.DB) *ActiveBlock {
	return &ActiveBlock{
		bloom:   metaBloom{hash: xxhash.New()},
		db:      db,
		entries: entry.NewMulti(),
	}
}

func (a *ActiveBlock) Save() error {
	a.mu.Lock()
	record := a.entries.Record()
	bloom := a.bloom.set(a.entries)
	a.entries.Reset()
	a.bloom.reset()
	a.mu.Unlock()
	return a.save(record, bloom)
}

func (a *ActiveBlock) Close() error {
	return nil
}

func (a *ActiveBlock) WriteEntry(e *entry.Entry) {
	a.mu.Lock()
	if a.Min.IsZero() {
		a.Min = e.Timestamp
	}
	a.Max = e.Timestamp
	a.entries.Append(e)
	a.mu.Unlock()
}

func (a *ActiveBlock) save(r arrow.Record, b *blocks.Bloom) error {
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
func Writer[T any](w io.Writer, o ...parquet.WriterOption) *parquet.GenericWriter[T] {
	var t T
	scheme := parquet.SchemaOf(t)
	var bloom []parquet.BloomFilterColumn
	for _, col := range scheme.Columns() {
		l, _ := scheme.Lookup(col...)
		if l.Node.Type().Kind() == parquet.ByteArray {
			bloom = append(bloom, parquet.SplitBlockFilter(FilterBitsPerValue, col...))
		}
	}
	return parquet.NewGenericWriter[T](w, parquet.BloomFilters(
		bloom...,
	))
}

type metaBloom struct {
	hash           *xxhash.Digest
	Browser        roaring64.Bitmap
	BrowserVersion roaring64.Bitmap
	City           roaring64.Bitmap
	Country        roaring64.Bitmap
	Domain         roaring64.Bitmap
	EntryPage      roaring64.Bitmap
	ExitPage       roaring64.Bitmap
	Host           roaring64.Bitmap
	Name           roaring64.Bitmap
	Os             roaring64.Bitmap
	OsVersion      roaring64.Bitmap
	Path           roaring64.Bitmap
	Referrer       roaring64.Bitmap
	ReferrerSource roaring64.Bitmap
	Region         roaring64.Bitmap
	Screen         roaring64.Bitmap
	UtmCampaign    roaring64.Bitmap
	UtmContent     roaring64.Bitmap
	UtmMedium      roaring64.Bitmap
	UtmSource      roaring64.Bitmap
	UtmTerm        roaring64.Bitmap
	UtmValue       roaring64.Bitmap
}

func (m *metaBloom) reset() {
	m.hash.Reset()
	m.Browser.Clear()
	m.BrowserVersion.Clear()
	m.City.Clear()
	m.Country.Clear()
	m.Domain.Clear()
	m.EntryPage.Clear()
	m.ExitPage.Clear()
	m.Host.Clear()
	m.Name.Clear()
	m.Os.Clear()
	m.OsVersion.Clear()
	m.Path.Clear()
	m.Referrer.Clear()
	m.ReferrerSource.Clear()
	m.Region.Clear()
	m.Screen.Clear()
	m.UtmCampaign.Clear()
	m.UtmContent.Clear()
	m.UtmMedium.Clear()
	m.UtmSource.Clear()
	m.UtmTerm.Clear()
	m.UtmValue.Clear()
}

func (m *metaBloom) sum(s string) uint64 {
	m.hash.Reset()
	m.hash.WriteString(s)
	return m.hash.Sum64()
}

func (m *metaBloom) ls(b *roaring64.Bitmap, values ...string) {
	for i := range values {
		if values[i] != "" {
			b.Add(m.sum(values[i]))
		}
	}
}

func (m *metaBloom) set(e *entry.MultiEntry) *blocks.Bloom {
	m.ls(&m.Browser, e.Browser...)
	m.ls(&m.BrowserVersion, e.BrowserVersion...)
	m.ls(&m.City, e.City...)
	m.ls(&m.Country, e.Country...)
	m.ls(&m.Domain, e.Domain...)
	m.ls(&m.EntryPage, e.EntryPage...)
	m.ls(&m.ExitPage, e.ExitPage...)
	m.ls(&m.Host, e.Host...)
	m.ls(&m.Name, e.Name...)
	m.ls(&m.Os, e.Os...)
	m.ls(&m.OsVersion, e.OsVersion...)
	m.ls(&m.Path, e.Path...)
	m.ls(&m.Referrer, e.Referrer...)
	m.ls(&m.Screen, e.Screen...)
	m.ls(&m.UtmCampaign, e.UtmCampaign...)
	m.ls(&m.UtmContent, e.UtmContent...)
	m.ls(&m.UtmMedium, e.UtmMedium...)
	m.ls(&m.UtmSource, e.UtmSource...)
	m.ls(&m.UtmTerm, e.UtmTerm...)
	return m.bloom()
}

func (m *metaBloom) bloom() (b *blocks.Bloom) {
	b = &blocks.Bloom{}
	if !m.Browser.IsEmpty() {
		b.Browser = must.Must(m.Browser.MarshalBinary())
	}
	if !m.BrowserVersion.IsEmpty() {
		b.BrowserVersion = must.Must(m.BrowserVersion.MarshalBinary())
	}
	if !m.City.IsEmpty() {
		b.City = must.Must(m.City.MarshalBinary())
	}
	if !m.Country.IsEmpty() {
		b.Country = must.Must(m.Country.MarshalBinary())
	}
	if !m.Domain.IsEmpty() {
		b.Domain = must.Must(m.Domain.MarshalBinary())
	}
	if !m.EntryPage.IsEmpty() {
		b.EntryPage = must.Must(m.EntryPage.MarshalBinary())
	}
	if !m.ExitPage.IsEmpty() {
		b.ExitPage = must.Must(m.ExitPage.MarshalBinary())
	}
	if !m.Host.IsEmpty() {
		b.Host = must.Must(m.Host.MarshalBinary())
	}
	if !m.Name.IsEmpty() {
		b.Name = must.Must(m.Name.MarshalBinary())
	}
	if !m.Os.IsEmpty() {
		b.Os = must.Must(m.Os.MarshalBinary())
	}
	if !m.OsVersion.IsEmpty() {
		b.OsVersion = must.Must(m.OsVersion.MarshalBinary())
	}
	if !m.Path.IsEmpty() {
		b.Path = must.Must(m.Path.MarshalBinary())
	}
	if !m.Referrer.IsEmpty() {
		b.Referrer = must.Must(m.Referrer.MarshalBinary())
	}
	if !m.ReferrerSource.IsEmpty() {
		b.ReferrerSource = must.Must(m.ReferrerSource.MarshalBinary())
	}
	if !m.Region.IsEmpty() {
		b.Region = must.Must(m.Region.MarshalBinary())
	}
	if !m.Screen.IsEmpty() {
		b.Screen = must.Must(m.Screen.MarshalBinary())
	}
	if !m.UtmCampaign.IsEmpty() {
		b.UtmCampaign = must.Must(m.UtmCampaign.MarshalBinary())
	}
	if !m.UtmContent.IsEmpty() {
		b.UtmContent = must.Must(m.UtmContent.MarshalBinary())
	}
	if !m.UtmMedium.IsEmpty() {
		b.UtmMedium = must.Must(m.UtmMedium.MarshalBinary())
	}
	if !m.UtmMedium.IsEmpty() {
		b.UtmMedium = must.Must(m.UtmMedium.MarshalBinary())
	}
	if !m.UtmSource.IsEmpty() {
		b.UtmSource = must.Must(m.Browser.MarshalBinary())
	}
	if !m.UtmTerm.IsEmpty() {
		b.UtmTerm = must.Must(m.UtmTerm.MarshalBinary())
	}
	if !m.UtmTerm.IsEmpty() {
		b.UtmTerm = must.Must(m.UtmTerm.MarshalBinary())
	}
	return
}
