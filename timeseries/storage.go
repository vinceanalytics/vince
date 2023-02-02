package timeseries

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/apache/arrow/go/v10/arrow"
	"github.com/apache/arrow/go/v10/arrow/array"
	"github.com/apache/arrow/go/v10/arrow/memory"
	"github.com/dgraph-io/badger/v3"
	"github.com/oklog/ulid/v2"
	"github.com/segmentio/parquet-go"
)

const (
	BucketPath      = "buckets"
	BucketMergePath = "merge"
	MetaPath        = "meta"
	ActiveFileName  = "active.parquet"
	SortRowCount    = int64(4089)
)

var ErrNoRows = errors.New("no rows")

type Storage[T any] struct {
	path       string
	activeFile *os.File
	writer     *parquet.SortingWriter[T]
	mu         sync.Mutex
	meta       *badger.DB
	pool       *sync.Pool
	allocator  memory.Allocator
}

func NewStorage[T any](path string) (*Storage[T], error) {
	dirs := []string{BucketPath, BucketMergePath, MetaPath}
	for _, p := range dirs {
		os.MkdirAll(filepath.Join(path, p), 0755)
	}
	db, err := badger.Open(badger.DefaultOptions(filepath.Join(path, MetaPath)))
	if err != nil {
		return nil, err
	}
	f, err := os.Create(filepath.Join(path, ActiveFileName))
	if err != nil {
		return nil, err
	}
	s := &Storage[T]{
		path:       path,
		activeFile: f,
		writer: parquet.NewSortingWriter[T](f, SortRowCount, parquet.SortingWriterConfig(
			parquet.SortingColumns(
				parquet.Ascending("timestamp"),
			),
		)),
		meta:      db,
		allocator: memory.DefaultAllocator,
	}
	s.pool = &sync.Pool{
		New: func() any {
			return s.build()
		},
	}
	return s, nil
}

func (s *Storage[T]) Write(rows []T) (int, error) {
	return s.writer.Write(rows)
}

func (s *Storage[T]) Archive() (int64, error) {
	return s.archive(true)
}

func (s *Storage[T]) archive(openActive bool) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.writer.Close()
	if err != nil {
		return 0, err
	}
	id := ulid.Make()
	ap := filepath.Join(s.path, BucketPath, id.String())
	f, err := os.Create(ap)
	if err != nil {
		return 0, err
	}
	s.activeFile.Seek(0, os.SEEK_SET)
	n, err := f.ReadFrom(s.activeFile)
	if err != nil {
		return 0, err
	}
	err = f.Close()
	if err != nil {
		return 0, err
	}
	err = s.meta.Update(func(txn *badger.Txn) error {
		return txn.Set(id.Bytes(), []byte{})
	})
	if err != nil {
		return 0, err
	}
	if openActive {
		a := filepath.Join(s.path, ActiveFileName)
		s.activeFile.Close()
		af, err := os.Create(a)
		if err != nil {
			return 0, err
		}
		s.activeFile = af
		s.writer.Reset(af)
	}
	return n, nil
}

func (s *Storage[T]) Close() error {
	_, err := s.archive(false)
	if err != nil {
		return err
	}
	return s.meta.Close()
}

func (s *Storage[T]) Query(from time.Time, to time.Time) error {
	ids, err := s.buckets(from, to)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return ErrNoRows
	}
	b := s.get()
	for _, id := range ids {
		err := b.processBucket(id, from, to)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage[T]) get() *StoreBuilder[T] {
	return s.pool.Get().(*StoreBuilder[T])
}

func (s *Storage[T]) put(b *StoreBuilder[T]) {
	s.pool.Put(b)
}

// returns all buckets which might contain series between from and to
func (s *Storage[T]) buckets(from time.Time, to time.Time) (ids []string, err error) {
	min := ulid.Timestamp(from)
	max := ulid.Timestamp(to)
	err = s.meta.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		for it.Rewind(); it.Valid(); it.Next() {
			key := it.Item().Key()
			var id ulid.ULID
			copy(id[:], key)
			ts := id.Time()
			if min <= ts && ts <= max {
				ids = append(ids, string(key))
			}
		}
		return nil
	})
	return
}

func (s *Storage[T]) build() *StoreBuilder[T] {
	fields := s.writer.Schema().Fields()
	b := &StoreBuilder[T]{
		store:   s,
		writers: make([]*Writer, len(fields)),
	}
	for i, f := range fields {
		dt, err := ParquetNodeToType(f)
		if err != nil {
			panic(err.Error())
		}
		b.writers[i] = &Writer{
			build: array.NewBuilder(s.allocator, dt),
			index: i,
			name:  f.Name(),
		}
	}
	b.active = append(b.active, b.writers[0])
	return b
}

type Writer struct {
	build array.Builder
	index int
	name  string
}

type StoreBuilder[T any] struct {
	store   *Storage[T]
	writers []*Writer
	active  []*Writer
}

func (b *StoreBuilder[T]) processBucket(id string, from time.Time, to time.Time) error {
	f, err := os.Open(filepath.Join(b.store.path, BucketPath, id))
	if err != nil {
		return err
	}
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	file, err := parquet.OpenFile(f, stat.Size())
	if err != nil {
		return err
	}
	for _, rg := range file.RowGroups() {
		if !b.take(rg) {
			continue
		}
	}
	return nil
}

func (b *StoreBuilder[T]) take(rg parquet.RowGroup) bool {
	ts := rg.ColumnChunks()[0]
	pages := ts.Pages()
	defer pages.Close()
	for {
		pg, err := pages.ReadPage()
		if err != nil {
			return false
		}
		min, max, ok := pg.Bounds()
		if !ok {
			return false
		}
		fmt.Printf("%#v %#v\n", min, max)
		return true
	}
}

func (b *StoreBuilder[T]) selectFields(names ...string) *StoreBuilder[T] {
	for _, w := range b.writers[1:] {
		for _, name := range names {
			if name == w.name {
				b.active = append(b.active, w)
				break
			}
		}
	}
	return b
}

func (b *StoreBuilder[T]) record() arrow.Record {
	cols := make([]arrow.Array, len(b.active))
	rows := int64(0)
	defer func() {
		for _, col := range cols {
			if col == nil {
				continue
			}
			col.Release()
		}
	}()
	fields := make([]arrow.Field, len(b.active))
	ps := b.store.writer.Schema().Fields()
	for i, w := range b.active {
		cols[i] = w.build.NewArray()
		rows += int64(cols[i].Len())
		pf := ps[w.index]
		fields[i] = arrow.Field{
			Name:     pf.Name(),
			Type:     w.build.Type(),
			Nullable: pf.Optional(),
		}
	}
	return array.NewRecord(arrow.NewSchema(fields, nil), cols, rows)
}
