package neo

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/parquet"
	"github.com/apache/arrow/go/v13/parquet/file"
	"github.com/apache/arrow/go/v13/parquet/pqarrow"
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/blocks"
	"github.com/vinceanalytics/vince/pkg/entry"
	"google.golang.org/protobuf/proto"
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

func (a *ActiveBlock) Save(ctx context.Context) error {
	ts := core.Now(ctx).UnixMilli()
	a.mu.Lock()
	if len(a.entries.Timestamp) == 0 {
		a.mu.Unlock()
		return nil
	}

	record := a.entries.Record(ts)
	bloom := a.bloom.set(a.entries)
	a.entries.Reset()
	a.bloom.reset()
	a.mu.Unlock()
	return a.save(ctx, ts, record, bloom)
}

func (a *ActiveBlock) Close() error {
	return a.Save(context.Background())
}

func (a *ActiveBlock) WriteEntry(e *entry.Entry) {
	a.mu.Lock()
	if a.Min.IsZero() {
		a.Min = e.Timestamp
	}
	a.Max = e.Timestamp
	a.entries.Append(e)
	a.mu.Unlock()
	e.Release()
}

func (a *ActiveBlock) save(ctx context.Context, ts int64, r arrow.Record, b *blocks.Bloom) (err error) {
	txn := a.db.NewTransaction(true)
	meta := must.Must(ReadMetadata(txn))
	var block *blocks.Block
	if len(meta.Blocks) > 0 {
		last := meta.Blocks[len(meta.Blocks)-1]
		if last.Size < (1 << 20) {
			block = last
			union(block.Bloom, b)
		}
	}
	if block == nil {
		id := ulid.Make()
		block = &blocks.Block{
			Id:    id.Bytes(),
			Min:   ts,
			Bloom: b,
		}
		meta.Blocks = append(meta.Blocks, block)
	}
	block.Max = ts
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
		r.Release()
	}()
	must.Assert(WriteBlock(ctx, txn, buf, block.Id, r))
	block.Size = int64(buf.Len())
	return errors.Join(
		txn.Set(metaPath, must.Must(proto.Marshal(meta))),
		must.Assert(txn.Commit()),
	)
}

func ReadMetadata(txn *badger.Txn) (*blocks.Metadata, error) {
	it, err := txn.Get(metaPath)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		return &blocks.Metadata{}, nil
	}
	meta := &blocks.Metadata{}
	err = it.Value(func(val []byte) error {
		return proto.Unmarshal(val, meta)
	})
	if err != nil {
		return nil, err
	}
	return meta, nil
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
	b = &blocks.Bloom{
		Filters: make(map[string][]byte),
	}
	if !m.Browser.IsEmpty() {
		b.Filters["browser"] = must.Must(m.Browser.MarshalBinary())
	}
	if !m.BrowserVersion.IsEmpty() {
		b.Filters["browser_version"] = must.Must(m.BrowserVersion.MarshalBinary())
	}
	if !m.City.IsEmpty() {
		b.Filters["city"] = must.Must(m.City.MarshalBinary())
	}
	if !m.Country.IsEmpty() {
		b.Filters["country"] = must.Must(m.Country.MarshalBinary())
	}
	if !m.Domain.IsEmpty() {
		b.Filters["domain"] = must.Must(m.Domain.MarshalBinary())
	}
	if !m.EntryPage.IsEmpty() {
		b.Filters["entry_page"] = must.Must(m.EntryPage.MarshalBinary())
	}
	if !m.ExitPage.IsEmpty() {
		b.Filters["exit_page"] = must.Must(m.ExitPage.MarshalBinary())
	}
	if !m.Host.IsEmpty() {
		b.Filters["host"] = must.Must(m.Host.MarshalBinary())
	}
	if !m.Name.IsEmpty() {
		b.Filters["name"] = must.Must(m.Name.MarshalBinary())
	}
	if !m.Os.IsEmpty() {
		b.Filters["os"] = must.Must(m.Os.MarshalBinary())
	}
	if !m.OsVersion.IsEmpty() {
		b.Filters["os_version"] = must.Must(m.OsVersion.MarshalBinary())
	}
	if !m.Path.IsEmpty() {
		b.Filters["path"] = must.Must(m.Path.MarshalBinary())
	}
	if !m.Referrer.IsEmpty() {
		b.Filters["referrer"] = must.Must(m.Referrer.MarshalBinary())
	}
	if !m.ReferrerSource.IsEmpty() {
		b.Filters["referrer_source"] = must.Must(m.ReferrerSource.MarshalBinary())
	}
	if !m.Region.IsEmpty() {
		b.Filters["region"] = must.Must(m.Region.MarshalBinary())
	}
	if !m.Screen.IsEmpty() {
		b.Filters["screen"] = must.Must(m.Screen.MarshalBinary())
	}
	if !m.UtmCampaign.IsEmpty() {
		b.Filters["utm_campaign"] = must.Must(m.UtmCampaign.MarshalBinary())
	}
	if !m.UtmContent.IsEmpty() {
		b.Filters["utm_content"] = must.Must(m.UtmContent.MarshalBinary())
	}
	if !m.UtmMedium.IsEmpty() {
		b.Filters["utm_medium"] = must.Must(m.UtmMedium.MarshalBinary())
	}

	if !m.UtmSource.IsEmpty() {
		b.Filters["utm_source"] = must.Must(m.Browser.MarshalBinary())
	}
	if !m.UtmTerm.IsEmpty() {
		b.Filters["utm_term"] = must.Must(m.UtmTerm.MarshalBinary())
	}
	return
}

func union(dst, src *blocks.Bloom) {
	var x, y roaring64.Bitmap
	for k, v := range src.Filters {
		h, ok := dst.Filters[k]
		if !ok {
			dst.Filters[k] = v
			continue
		}
		x.Clear()
		x.UnmarshalBinary(h)

		y.Clear()
		y.UnmarshalBinary(v)

		x.Or(&y)
		dst.Filters[k] = must.Must(x.MarshalBinary())
	}
}

// WriteBlock saves record r in a parquet file with key. If the block exists a
// new file is created that adds record r to it.
func WriteBlock(ctx context.Context, txn *badger.Txn, b *bytes.Buffer, key []byte, r arrow.Record) (err error) {
	it, err := txn.Get(key)
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		w, err := pqarrow.NewFileWriter(entry.Schema, b,
			parquet.NewWriterProperties(
				parquet.WithAllocator(entry.Pool),
			),
			pqarrow.NewArrowWriterProperties(
				pqarrow.WithAllocator(entry.Pool),
			))

		if err != nil {
			return err
		}
		err = w.Write(r)
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
		return txn.Set(key, b.Bytes())
	}
	err = it.Value(func(val []byte) error {
		pr, err := pqarrow.ReadTable(ctx, bytes.NewReader(val), &parquet.ReaderProperties{}, pqarrow.ArrowReadProperties{
			Parallel: true,
		}, entry.Pool)
		if err != nil {
			return err
		}
		defer pr.Release()
		w, err := pqarrow.NewFileWriter(entry.Schema, b,
			parquet.NewWriterProperties(
				parquet.WithAllocator(entry.Pool),
			),
			pqarrow.NewArrowWriterProperties(
				pqarrow.WithAllocator(entry.Pool),
			))
		if err != nil {
			return err
		}
		err = w.WriteTable(pr, 1<<20)
		if err != nil {
			return err
		}
		err = w.Write(r)
		if err != nil {
			return err
		}
		return w.Close()
	})
	if err != nil {
		return err
	}
	return txn.Set(key, b.Bytes())
}

var bufferPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// ReadBlock reads records constructed by combining fields from block with key.
// For each record cb is called with it. If cb returns false reading is halter.
func ReadBlock(ctx context.Context, db *badger.DB, key []byte, a Analysis) error {
	return db.View(func(txn *badger.Txn) error {
		it, err := txn.Get(key)
		if err != nil {
			return err
		}
		return it.Value(func(val []byte) error {
			f, err := file.NewParquetReader(bytes.NewReader(val),
				file.WithReadProps(parquet.NewReaderProperties(entry.Pool)))
			if err != nil {
				return err
			}
			r, err := pqarrow.NewFileReader(f, pqarrow.ArrowReadProperties{
				Parallel: true,
			}, entry.Pool)
			if err != nil {
				return err
			}
			x, err := r.GetRecordReader(ctx, a.ColumnIndices(), nil)
			if err != nil {
				return err
			}
			a.Analyze(ctx, x)
			x.Release()
			return nil
		})
	})
}
