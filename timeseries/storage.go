package timeseries

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/apache/arrow/go/v12/arrow"
	"github.com/apache/arrow/go/v12/arrow/array"
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
		allocator: memory.DefaultAllocator,
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

func (s *Storage[T]) Query(query Query) (*Record, error) {
	ids, err := s.meta.Buckets(query.start, query.end)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, ErrNoRows
	}
	b := s.get()
	defer s.put(b)
	for _, f := range query.filters {
		for _, w := range b.writers {
			if w.name == f.field {
				b.active = append(b.active, w)
				break
			}
		}
	}
	for _, id := range ids {
		err := b.processBucket(id, query)
		if err != nil {
			return nil, err
		}
	}
	return b.record()
}

func (s *Storage[T]) get() *StoreBuilder[T] {
	return s.pool.Get().(*StoreBuilder[T])
}

func (s *Storage[T]) put(b *StoreBuilder[T]) {
	b.active = b.active[:0]
	b.active = append(b.active, b.writers[0])
	for _, a := range b.results {
		a.Release()
	}
	b.results = b.results[:0]
	s.pool.Put(b)
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

func (w *Writer) WritePage(p parquet.Page, valid []bool) error {
	switch b := w.build.(type) {
	case *array.Int64Builder:
		r := p.Values().(parquet.Int64Reader)
		a := make([]int64, p.NumValues())
		if _, err := r.ReadInt64s(a); err != nil && !errors.Is(err, io.EOF) {
			return err
		} else {
			b.AppendValues(a, valid)
		}
	case *array.BinaryDictionaryBuilder:
	default:
		fmt.Printf("%#T %#T\n", w.build, p.Values())
	}
	return nil
}

type StoreBuilder[T any] struct {
	store   *Storage[T]
	writers []*Writer
	active  []*Writer
	results []arrow.Array
}

func (b *StoreBuilder[T]) processBucket(id string, query Query) error {
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
		err = b.RowGroup(rg, query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *StoreBuilder[T]) RowGroup(rg parquet.RowGroup, query Query) error {
	start := query.start.UnixNano()
	end := query.end.UnixNano()
	chunks := rg.ColumnChunks()
	cs := make([]parquet.Pages, len(b.active))
	for i, w := range b.active {
		cs[i] = chunks[w.index].Pages()
	}
	defer func() {
		for _, p := range cs {
			p.Close()
		}
	}()
	var ls [][]*PageIndex
	var i int
	for {
		p, err := cs[0].ReadPage()
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
		pg := addPageIndex(p, start, end, minValue, maxValue)
		if pg != nil {
			pls := make([]*PageIndex, len(b.active))
			pls[0] = pg
			for i, np := range cs[1:] {
				px, err := np.ReadPage()
				if err != nil {
					return err
				}
				pls[i+1] = &PageIndex{
					Page:    px,
					Partial: pg.Partial,
				}
			}
			ls = append(ls, pls)
		}
		i += 1
	}
	for _, pages := range ls {
		b.processPages(rg, pages, query)
	}
	return nil
}

func (b *StoreBuilder[T]) processPages(rg parquet.RowGroup, pages []*PageIndex, query Query) {
	filter := make([]bool, pages[0].Page.NumValues())
	if !b.filterPages(rg, pages, filter, query) {
		// if any filter decided we should skip this page we skip it right away
		return
	}
	b.active[0].WritePage(pages[0].Page, filter)
	b.results = append(b.results, b.active[0].build.NewArray())
}

func (b *StoreBuilder[T]) filterPages(rg parquet.RowGroup, pages []*PageIndex, filter []bool, query Query) bool {
	for i, p := range pages[1:] {
		a := b.active[i+1]
		for _, f := range query.filters {
			if a.name == f.field {
				if f.h != nil {
					if !f.h(filter, a.index, rg, p.Page) {
						return false
					}
				}
			}
		}
	}
	return true
}

func addPageIndex(page parquet.Page, start, end, min, max int64) *PageIndex {
	if start <= min && end <= max {
		return &PageIndex{
			Page: page,
		}
	} else if end <= max {
		return &PageIndex{
			Page:    page,
			Partial: PartialMax,
		}
	} else if start <= min {
		return &PageIndex{
			Page:    page,
			Partial: PartialMin,
		}
	}
	return nil
}

type PageIndex struct {
	Page    parquet.Page
	Index   int
	Partial PageIndexState
}

type PageIndexState uint

const (
	Full PageIndexState = iota
	PartialMax
	PartialMin
)

type Record struct {
	Labels     map[string]string `json:"labels,omitempty"`
	Timestamps arrow.Array       `json:"timestamps,omitempty"`
}

func (b *StoreBuilder[T]) record() (*Record, error) {
	ts, err := array.Concatenate(b.results, b.store.allocator)
	if err != nil {
		return nil, err
	}
	return &Record{
		Timestamps: ts,
	}, nil
}
