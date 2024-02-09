package db

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/ipc"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/index"
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

func (s *Store) Save(r arrow.Record, idx index.Full) (*v1.Granule, error) {

	// We don't call this frequently. So make sure we run GC when we are done. This
	// removes the need for periodic GC calls.
	defer s.db.GC()

	id := ulid.Make().String()
	var key bytes.Buffer
	key.WriteString(s.resource)
	key.Write(slash)
	key.WriteString(id)
	base := bytes.Clone(key.Bytes())
	size, err := s.SaveRecord(&key, base, r)
	if err != nil {
		return nil, err
	}
	var kb bytes.Buffer
	idx.Columns(func(column index.Column) error {
		if column.Empty() {
			// skip empty indexes.
			return nil
		}
		kb.Reset()
		for _, p := range column.Path() {
			kb.Write(slash)
			kb.WriteString(p)
		}
		return s.SaveIndex(&key, base, kb.Bytes(), column)
	})
	lo, hi := Timestamps(r)
	return &v1.Granule{
		Min:  lo,
		Max:  hi,
		Size: size + idx.Size(),
		Id:   id,
		Rows: uint64(r.NumRows()),
	}, nil
}

func (s *Store) SaveRecord(
	buf *bytes.Buffer,
	base []byte,
	r arrow.Record,
) (n uint64, err error) {
	schema := r.Schema()
	var x uint64
	for i := 0; i < int(r.NumCols()); i++ {
		x, err = s.SaveColumn(buf, base, r.ColumnName(i), schema.Field(i), r.Column(i))
		if err != nil {
			return
		}
		n += x
	}
	return
}

func (s *Store) SaveColumn(
	buf *bytes.Buffer,
	base []byte,
	key string,
	field arrow.Field,
	a arrow.Array,
) (n uint64, err error) {
	r := array.NewRecord(
		arrow.NewSchema([]arrow.Field{field}, nil),
		[]arrow.Array{a},
		int64(a.Len()),
	)
	defer r.Release()
	b := persistBuffer.Get().(*bytes.Buffer)
	defer func() {
		n = uint64(b.Len())
		b.Reset()
		persistBuffer.Put(b)
	}()
	w := ipc.NewWriter(b,
		ipc.WithSchema(r.Schema()),
		ipc.WithAllocator(s.mem),
		ipc.WithZstd(),
		ipc.WithMinSpaceSavings(0.3), //at least 30% savings
	)
	err = w.Write(r)
	if err != nil {
		return
	}
	err = w.Close()
	if err != nil {
		return
	}
	buf.Reset()
	buf.Write(base)
	buf.Write(slash)
	buf.Write(recordBytes)
	buf.Write(slash)
	buf.WriteString(key)
	err = s.db.Set(buf.Bytes(), b.Bytes(), s.ttl)
	return
}

func (s *Store) SaveIndex(
	buf *bytes.Buffer,
	base []byte,
	key []byte,
	idx index.Column,
) error {
	buf.Reset()
	buf.Write(base)
	buf.Write(slash)
	buf.Write(fstBytes)
	buf.Write(key) // [resource/id/fst/key]
	err := s.db.Set(buf.Bytes(), idx.Fst(), s.ttl)
	if err != nil {
		return err
	}
	buf.Reset()
	buf.Write(base)
	buf.Write(slash)
	buf.Write(bitmapBytes)

	buf.Write([]byte(key))
	buf.Write(slash)

	// [resource/id/bitmaps/key]
	base = bytes.Clone(buf.Bytes())
	return idx.Bitmaps(func(i int, b *roaring.Bitmap) error {
		buf.Reset()
		buf.Write(base)
		fmt.Fprint(buf, i) // [resource/id/bitmaps/key/row]
		data, err := b.MarshalBinary()
		if err != nil {
			return err
		}
		return s.db.Set(buf.Bytes(), data, s.ttl)
	})
}

var (
	slash       = []byte("/")
	fstBytes    = []byte("fst")
	bitmapBytes = []byte("bitmap")
	recordBytes = []byte("record")
)

var persistBuffer = &sync.Pool{New: func() any { return new(bytes.Buffer) }}

type KV struct {
	db *badger.DB
}

type dbKey struct{}

func With(ctx context.Context, kv Storage) context.Context {
	return context.WithValue(ctx, dbKey{}, kv)
}

func Get(ctx context.Context) Storage {
	return ctx.Value(dbKey{}).(Storage)
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
