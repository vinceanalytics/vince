package timeseries

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/apache/arrow/go/v12/arrow"
	"github.com/apache/arrow/go/v12/arrow/array"
	"github.com/apache/arrow/go/v12/arrow/compute"
	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/dgraph-io/badger/v3"
	"github.com/oklog/ulid/v2"
	"github.com/segmentio/parquet-go"
)

//go:generate protoc -I=. --go_out=paths=source_relative:. storage_schema.proto

const (
	BucketPath       = "buckets"
	BucketMergePath  = "merge"
	MetaPath         = "meta"
	ActiveFileName   = "active.parquet"
	RealTimeFileName = "realtime.parquet"
	SortRowCount     = int64(4089)
)

var ErrNoRows = errors.New("no rows")

type Storage[T any] struct {
	path       string
	activeFile *os.File
	writer     *parquet.SortingWriter[T]
	mu         sync.Mutex
	meta       *meta
	pool       *sync.Pool
	allocator  memory.Allocator
	start      time.Time
	end        time.Time
}

func NewStorage[T any](allocator memory.Allocator, path string) (*Storage[T], error) {
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
	now := time.Now()
	s := &Storage[T]{
		path:       path,
		activeFile: f,
		writer: parquet.NewSortingWriter[T](f, SortRowCount, parquet.SortingWriterConfig(
			parquet.SortingColumns(
				parquet.Ascending("timestamp"),
			),
		)),
		meta:      &meta{db: db},
		allocator: allocator,
		start:     now,
		end:       now,
	}
	s.pool = &sync.Pool{
		New: func() any {
			return s.build()
		},
	}
	return s, nil
}

func (s *Storage[T]) Write(lastTimestamp time.Time, rows []T) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.end = lastTimestamp
	return s.writer.Write(rows)
}

func (s *Storage[T]) Archive() (int64, error) {
	return s.archive(true)
}

func (s *Storage[T]) archive(openActive bool) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.start.Equal(s.end) {
		return 0, nil
	}
	err := s.writer.Close()
	if err != nil {
		return 0, err
	}
	id := ulid.Make()

	n, err := createFile(filepath.Join(s.path, BucketPath, id.String()), s.activeFile)
	if err != nil {
		return 0, err
	}
	_, err = createAtomicFile(filepath.Join(s.path, RealTimeFileName), s.activeFile)
	if err != nil {
		return 0, err
	}
	err = s.meta.SaveBucket(id, s.start, s.end)
	if err != nil {
		return 0, err
	}
	s.start = s.end
	s.end = s.start
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

func createFile(out string, src *os.File) (int64, error) {
	f, err := os.Create(out)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	src.Seek(0, io.SeekStart)
	return f.ReadFrom(src)
}

func createAtomicFile(out string, src *os.File) (n int64, err error) {
	f, err := ioutil.TempFile(filepath.Dir(out), filepath.Base(out))
	if err != nil {
		return 0, err
	}
	defer func() {
		os.Remove(f.Name())
	}()
	src.Seek(0, io.SeekStart)
	n, err = f.ReadFrom(src)
	if err != nil {
		return
	}
	err = f.Close()
	if err != nil {
		return
	}
	err = os.Rename(f.Name(), out)
	return
}

func (s *Storage[T]) Close() error {
	_, err := s.archive(false)
	if err != nil {
		return err
	}
	return s.meta.Close()
}

func (s *Storage[T]) Query(ctx context.Context, query Query) (*Record, error) {
	ids, err := s.meta.Buckets(query.start, query.end)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, ErrNoRows
	}
	files := make([]string, len(ids))
	for i := range ids {
		files[i] = filepath.Join(s.path, BucketPath, ids[i])
	}
	return s.query(ctx, query, files...)
}

func (s *Storage[T]) QueryRealtime(ctx context.Context, query Query) (*Record, error) {
	return s.query(ctx, query, filepath.Join(s.path, RealTimeFileName))
}

func (s *Storage[T]) query(ctx context.Context, query Query, files ...string) (*Record, error) {
	b := s.get()
	defer s.put(b)
	m := make(map[string]*Writer)
	for _, w := range b.writers {
		m[w.name] = w
	}
	for _, n := range query.selected {
		if w, ok := m[n]; ok {
			b.active = append(b.active, w)
			b.selected = append(b.selected, w)
		}
	}
	for _, f := range query.filters {
		if w, ok := m[f.field]; ok {
			b.active = append(b.active, w)
		}
	}
	for _, file := range files {
		err := b.processFile(ctx, file, query)
		if err != nil {
			return nil, err
		}
	}
	return b.record(ctx)
}

func (s *Storage[T]) get() *StoreBuilder[T] {
	return s.pool.Get().(*StoreBuilder[T])
}

func (s *Storage[T]) put(b *StoreBuilder[T]) {
	b.reset()
	s.pool.Put(b)
}

func (s *Storage[T]) build() *StoreBuilder[T] {
	fields := s.writer.Schema().Fields()
	b := &StoreBuilder[T]{
		store:   s,
		writers: make([]*Writer, len(fields)),
		filter:  array.NewBooleanBuilder(s.allocator),
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

func (w *Writer) WritePage(p parquet.Page) error {
	switch b := w.build.(type) {
	case *array.Int64Builder:
		r := p.Values().(parquet.Int64Reader)
		a := make([]int64, p.NumValues())
		if _, err := r.ReadInt64s(a); err != nil && !errors.Is(err, io.EOF) {
			return err
		} else {
			b.AppendValues(a, nil)
		}
	default:
	}
	return nil
}

type StoreBuilder[T any] struct {
	store    *Storage[T]
	writers  []*Writer
	active   []*Writer
	selected []*Writer
	filter   *array.BooleanBuilder
}

func (b *StoreBuilder[T]) reset() {
	b.active = b.active[:0]
	b.active = append(b.active, b.writers[0])
	b.selected = b.selected[:0]
}

func (b *StoreBuilder[T]) processFile(ctx context.Context, filePath string, query Query) error {
	f, err := os.Open(filePath)
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
		err = b.RowGroup(ctx, rg, query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *StoreBuilder[T]) RowGroup(ctx context.Context, rg parquet.RowGroup, query Query) error {
	start := query.start.UnixNano()
	end := query.end.UnixNano()
	chunks := rg.ColumnChunks()
	cs := make(map[string]parquet.Pages)

	for _, w := range b.active {
		cs[w.name] = chunks[w.index].Pages()
	}
	defer func() {
		for _, p := range cs {
			p.Close()
		}
	}()
	var ls [][]parquet.Page
	for {
		p, err := cs[b.active[0].name].ReadPage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		min, max, ok := p.Bounds()
		if !ok {
			break
		}
		minValue := min.Int64()
		maxValue := max.Int64()
		if start <= minValue && end <= maxValue {
			pls := make([]parquet.Page, len(b.active))
			pls[0] = p
			for i, w := range b.active[1:] {
				px, err := cs[w.name].ReadPage()
				if err != nil {
					return err
				}
				pls[i+1] = px
			}
			ls = append(ls, pls)
		}
	}
	for _, pages := range ls {
		b.processPages(ctx, pages, query)
	}
	return nil
}

func (b *StoreBuilder[T]) processPages(ctx context.Context, pages []parquet.Page, query Query) {
	filter := make([]bool, pages[0].NumValues())
	if !b.filterPages(ctx, pages, filter, query) {
		// if any filter decided we should skip this page we skip it right away
		return
	}
	b.filter.AppendValues(filter, nil)
	for i, s := range b.selected {
		// selected fields starts from index 1 of active fields
		s.WritePage(pages[i+1])
	}
}

func (b *StoreBuilder[T]) filterPages(ctx context.Context, pages []parquet.Page, filter []bool, query Query) bool {
	for i, p := range pages[1:] {
		a := b.active[i+1]
		for _, f := range query.filters {
			if a.name == f.field {
				if f.h != nil {
					if !f.h(ctx, p) {
						return false
					}
				}
			}
		}
	}
	return true
}

type Record struct {
	Labels map[string]string `json:"labels,omitempty"`
	Fields []*Field          `json:"fields,omitempty"`
}

type Field struct {
	Name  string      `json:"name"`
	Value arrow.Array `json:"value"`
}

func (b *StoreBuilder[T]) record(ctx context.Context) (*Record, error) {
	r := &Record{}
	filter := b.filter.NewArray()
	for _, s := range b.selected {
		a, err := compute.FilterArray(ctx, s.build.NewArray(), filter, *compute.DefaultFilterOptions())
		if err != nil {
			return nil, err
		}
		r.Fields = append(r.Fields, &Field{
			Name:  s.name,
			Value: a,
		})
	}
	return r, nil
}
