package db

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/apache/arrow/go/v15/parquet"
	"github.com/apache/arrow/go/v15/parquet/compress"
	"github.com/apache/arrow/go/v15/parquet/pqarrow"
	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/buffers"
	"github.com/vinceanalytics/vince/internal/index"
)

var ErrKeyNotFound = errors.New("db: key not found")

type Storage interface {
	Set(key, value []byte, ttl time.Duration) error
	Get(key []byte, value func([]byte) error) error
	GC() error
	Close() error
}

type Store struct {
	db  Storage
	mem memory.Allocator
	ttl time.Duration
	log *slog.Logger
}

func NewStore(db Storage, mem memory.Allocator, ttl time.Duration) *Store {
	return &Store{
		db:  db,
		ttl: ttl,
		mem: mem,
		log: slog.Default().With(
			slog.String("component", "lsm-store"),
		),
	}
}

func (s *Store) LoadRecord(ctx context.Context, resource, id string, numRows int64, columns []int) (arrow.Record, error) {
	return NewRecord(ctx, s.db, resource, id, numRows, columns)
}

func (s *Store) LoadIndex(ctx context.Context, resource, id string) (*index.FileIndex, error) {
	return NewIndex(ctx, s.db, resource, id)
}

func (s *Store) Save(resource string, r arrow.Record, idx index.Full) (*v1.Granule, error) {

	// We don't call this frequently. So make sure we run GC when we are done. This
	// removes the need for periodic GC calls.
	defer s.db.GC()

	id := ulid.Make().String()
	buf := buffers.Bytes()
	defer buf.Release()

	buf.WriteString(resource)
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

	return &v1.Granule{
		Min:  int64(idx.Min()),
		Max:  int64(idx.Max()),
		Size: size + uint64(buf.Len()),
		Id:   id,
		Rows: uint64(r.NumRows()),
	}, nil
}

func (s *Store) SaveRecord(buf *buffers.BytesBuffer, base []byte, r arrow.Record) (n uint64, err error) {
	b := buffers.Bytes()
	defer b.Release()

	err = ArrowToParquet(b, s.mem, r)
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

// We use zstd SpeedBestCompression because we rarely touch cold data and when
// touched, we heavily cache columns.
//
// We optimize for on disk size. It is better to try to fit granules in the lsm
// tree for faster lookups.
const compressionLevel = 11

func ArrowToParquet(out io.Writer, mem memory.Allocator, r arrow.Record) error {
	props := []parquet.WriterProperty{
		parquet.WithAllocator(mem),
		parquet.WithBatchSize(r.NumRows()), // we save as a single row group
		parquet.WithCompression(compress.Codecs.Zstd),
		parquet.WithCompressionLevel(compressionLevel),
	}
	for i := 0; i < int(r.NumCols()); i++ {
		if r.Column(i).DataType().ID() == arrow.DICTIONARY {
			props = append(props, parquet.WithDictionaryFor(r.ColumnName(i), true))
		}
	}
	w, err := pqarrow.NewFileWriter(r.Schema(), out, parquet.NewWriterProperties(props...),
		pqarrow.NewArrowWriterProperties(
			pqarrow.WithAllocator(mem),
			pqarrow.WithStoreSchema(),
		))
	if err != nil {
		return err
	}
	err = w.Write(r)
	if err != nil {
		return err
	}
	return w.Close()
}

var (
	slash    = []byte("/")
	dataPath = []byte("data")
)

type KV struct {
	DB *badger.DB
}

func OpenBadger(path string) (*badger.DB, error) {
	return badger.Open(badger.DefaultOptions(filepath.Join(path, "db")).
		WithLogger(&badgerLogger{
			log: slog.Default().With(
				slog.String("component", "key-value-store"),
			),
		}))
}

func NewKV(path string) (*KV, error) {
	db, err := OpenBadger(path)
	if err != nil {
		return nil, err
	}
	return &KV{DB: db}, nil
}

var _ Storage = (*KV)(nil)

func (kv *KV) GC() error {
	return kv.DB.RunValueLogGC(0.5)
}

func (kv *KV) Close() error {
	return kv.DB.Close()
}

func (kv *KV) Set(key, value []byte, ttl time.Duration) error {
	println("=>", string(key))
	e := badger.NewEntry(key, value)
	if ttl > 0 {
		e = e.WithTTL(ttl)
	}
	return kv.DB.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(e)
	})
}

func (kv *KV) Get(key []byte, value func([]byte) error) error {
	return kv.DB.View(func(txn *badger.Txn) error {
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
