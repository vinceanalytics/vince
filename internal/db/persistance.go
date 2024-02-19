package db

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/apache/arrow/go/v15/parquet"
	"github.com/apache/arrow/go/v15/parquet/compress"
	"github.com/apache/arrow/go/v15/parquet/pqarrow"
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/internal/buffers"
	"github.com/vinceanalytics/vince/internal/columns"
	"github.com/vinceanalytics/vince/internal/index"
	"github.com/vinceanalytics/vince/internal/logger"
)

var ErrKeyNotFound = errors.New("db: key not found")

type Storage interface {
	Set(key, value []byte, ttl time.Duration) error
	Get(key []byte, value func([]byte) error) error
	GC() error
	Close() error
}

type Store struct {
	db       Storage
	mem      memory.Allocator
	ttl      time.Duration
	resource string
	log      *slog.Logger
}

func NewStore(db Storage, mem memory.Allocator, resource string, ttl time.Duration) *Store {
	return &Store{
		db:       db,
		ttl:      ttl,
		resource: resource,
		mem:      mem,
		log: slog.Default().With(
			slog.String("component", "lsm-store"),
			slog.String("resource", resource),
		),
	}
}

func (s *Store) LoadRecord(ctx context.Context, id string, numRows int64, columns []int) (arrow.Record, error) {
	return NewRecord(ctx, s.db, s.resource, id, numRows, columns)
}

func (s *Store) LoadIndex(ctx context.Context, id string) (*index.FileIndex, error) {
	return NewIndex(ctx, s.db, s.resource, id)
}

func (s *Store) Save(r arrow.Record, idx index.Full) (*v1.Granule, error) {

	// We don't call this frequently. So make sure we run GC when we are done. This
	// removes the need for periodic GC calls.
	defer s.db.GC()

	id := ulid.Make().String()
	buf := buffers.Bytes()
	defer buf.Release()

	buf.WriteString(s.resource)
	buf.Write(slash)
	buf.WriteString(id)
	base := bytes.Clone(buf.Bytes())
	size, err := s.SaveRecord(buf, base, r)
	if err != nil {
		return nil, err
	}
	buf.Reset()
	err = index.WriteFull(buf, idx, id)
	if err != nil {
		return nil, err
	}
	err = s.db.Set(base, buf.Bytes(), s.ttl)
	if err != nil {
		return nil, err
	}
	bm := new(roaring64.Bitmap)
	for i := 0; i < int(r.NumCols()); i++ {
		if r.ColumnName(i) == columns.Domain {
			h := new(xxhash.Digest)
			a := r.Column(i).(*array.Dictionary).Dictionary().(*array.String)
			for n := 0; n < a.Len(); n++ {
				h.WriteString(a.Value(i))
				bm.Add(h.Sum64())
				h.Reset()
			}
			break
		}
	}
	sites, err := bm.MarshalBinary()
	if err != nil {
		logger.Fail("Failed marshalling sites bitmap", "err", err)
	}
	return &v1.Granule{
		Min:   int64(idx.Min()),
		Max:   int64(idx.Max()),
		Size:  size + uint64(buf.Len()),
		Id:    id,
		Rows:  uint64(r.NumRows()),
		Sites: sites,
	}, nil
}

func (s *Store) SaveRecord(buf *buffers.BytesBuffer, base []byte, r arrow.Record) (n uint64, err error) {
	schema := r.Schema()
	b := buffers.Bytes()
	defer b.Release()
	props := []parquet.WriterProperty{
		parquet.WithAllocator(s.mem),
		parquet.WithBatchSize(r.NumRows()), // we save as a single row group
		parquet.WithCompression(compress.Codecs.Zstd),
	}
	for i := 0; i < int(r.NumCols()); i++ {
		if r.Column(i).DataType().ID() == arrow.DICTIONARY {
			props = append(props, parquet.WithDictionaryFor(r.ColumnName(i), true))
		}
	}
	w, err := pqarrow.NewFileWriter(schema, b, parquet.NewWriterProperties(props...),
		pqarrow.NewArrowWriterProperties(
			pqarrow.WithAllocator(s.mem),
			pqarrow.WithStoreSchema(),
		))
	if err != nil {
		return 0, err
	}
	err = w.Write(r)
	if err != nil {
		return 0, err
	}
	err = w.Close()
	if err != nil {
		return 0, err
	}
	buf.Reset()
	buf.Write(base)
	buf.Write(slash)
	buf.Write(dataPath)
	err = s.db.Set(buf.Bytes(), b.Bytes(), s.ttl)
	n = uint64(b.Len())
	return
}

var (
	slash    = []byte("/")
	dataPath = []byte("data")
)

type KV struct {
	db *badger.DB
}

func NewKV(path string) (*KV, error) {
	db, err := badger.Open(badger.DefaultOptions(filepath.Join(path, "db")).
		WithLogger(&badgerLogger{
			log: slog.Default().With(
				slog.String("component", "key-value-store"),
			),
		}))
	if err != nil {
		return nil, err
	}
	return &KV{db: db}, nil
}

var _ Storage = (*KV)(nil)

func (kv *KV) GC() error {
	return kv.db.RunValueLogGC(0.5)
}

func (kv *KV) Close() error {
	return kv.db.Close()
}

func (kv *KV) Set(key, value []byte, ttl time.Duration) error {
	println("=>", string(key))
	e := badger.NewEntry(key, value)
	if ttl > 0 {
		e = e.WithTTL(ttl)
	}
	return kv.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(e)
	})
}

func (kv *KV) Get(key []byte, value func([]byte) error) error {
	return kv.db.View(func(txn *badger.Txn) error {
		it, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrKeyNotFound
			}
			return err
		}
		return it.Value(value)
	})
}

type badgerLogger struct {
	log *slog.Logger
}

var _ badger.Logger = (*badgerLogger)(nil)

func (b *badgerLogger) Errorf(msg string, args ...interface{}) {
	b.log.Error(fmt.Sprintf(msg, args...))
}
func (b *badgerLogger) Warningf(msg string, args ...interface{}) {
	b.log.Warn(fmt.Sprintf(msg, args...))
}
func (b *badgerLogger) Infof(msg string, args ...interface{}) {
	b.log.Info(fmt.Sprintf(msg, args...))
}
func (b *badgerLogger) Debugf(msg string, args ...interface{}) {
	b.log.Debug(fmt.Sprintf(msg, args...))
}

type PrefixStore struct {
	Storage
	prefix []byte
}

func NewPrefix(store Storage, prefix string) Storage {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return &PrefixStore{
		Storage: store,
		prefix:  []byte(prefix),
	}
}

func (p *PrefixStore) Set(key, value []byte, ttl time.Duration) error {
	return p.Storage.Set(p.key(key), value, ttl)
}

func (p *PrefixStore) Get(key []byte, value func([]byte) error) error {
	return p.Storage.Get(p.key(key), value)
}

func (p *PrefixStore) key(k []byte) []byte {
	return append(p.prefix, k...)
}
