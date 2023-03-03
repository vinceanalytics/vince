package timeseries

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/apache/arrow/go/v12/arrow"
	"github.com/apache/arrow/go/v12/arrow/array"
	"github.com/apache/arrow/go/v12/arrow/compute"
	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/gernest/vince/log"
	"github.com/segmentio/parquet-go"
)

const (
	SortRowCount = int64(4089)
)

var ErrNoRows = errors.New("no rows")
var ErrSkipPage = errors.New("skip page")

type StoreItem interface {
	*Event | *Session
}

func QueryTable[T StoreItem](ctx context.Context, model T, uid uint64, query Query, files ...string) (*Record, error) {
	bob := Bob{db: Get(ctx).db}
	t := any(model)
	var b *StoreBuilder[T]
	var table TableID
	switch t.(type) {
	case *Event:
		table = EVENTS
		b = eventsPool.Get().(*StoreBuilder[T])
		defer func() {
			b.reset()
			eventsPool.Put(b)
		}()
	case *Session:
		table = SESSIONS
		b = sessionsPool.Get().(*StoreBuilder[T])
		defer func() {
			b.reset()
			sessionsPool.Put(b)
		}()
	}
	m := make(map[string]*Writer)
	set := make(map[string]bool)
	for _, w := range b.writers {
		m[w.name] = w
	}
	for _, n := range query.selected {
		if w, ok := m[n]; ok {
			b.active = append(b.active, w)
			b.selected = append(b.selected, w)
			set[n] = true
		}
	}
	for _, f := range query.filters {
		if w, ok := m[f.field]; ok && !set[f.field] {
			b.active = append(b.active, w)
			set[f.field] = true
		}
	}
	if cap(b.results) < len(b.selected) {
		b.results = make([][]arrow.Array, len(b.selected))
	}
	b.results = b.results[0:len(b.selected)]

	err := bob.Iterate(ctx, table, uid, query.start, query.end, b.process(ctx, query))
	if err != nil {
		return nil, err
	}
	return b.record(ctx)
}

var eventsPool = &sync.Pool{
	New: func() any {
		return build(&Event{})
	},
}

var sessionsPool = &sync.Pool{
	New: func() any {
		return build(&Session{})
	},
}

func build[T StoreItem](model T) *StoreBuilder[T] {
	fields := parquet.SchemaOf(model).Fields()
	b := &StoreBuilder[T]{
		writers: make([]*Writer, len(fields)),
		builders: &Builders{
			Int64:  array.NewInt64Builder(memory.DefaultAllocator),
			String: array.NewStringBuilder(memory.DefaultAllocator),
			Bool:   array.NewBooleanBuilder(memory.DefaultAllocator),
		},
	}
	for i, f := range fields {
		dt, err := ParquetNodeToType(f)
		if err != nil {
			panic(err.Error())
		}
		b.writers[i] = &Writer{
			build: array.NewBuilder(memory.DefaultAllocator, dt),
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
	case *array.StringBuilder:
		a := make([]parquet.Value, p.NumValues())
		if _, err := p.Values().ReadValues(a); err != nil && !errors.Is(err, io.EOF) {
			return err
		} else {
			for i := 0; i < int(p.NumValues()); i += 1 {
				b.Append(a[i].String())
			}
		}
	default:
	}
	return nil
}

type StoreBuilder[T any] struct {
	writers  []*Writer
	active   []*Writer
	selected []*Writer
	results  [][]arrow.Array
	builders *Builders
}

func (b *StoreBuilder[T]) reset() {
	b.active = b.active[:0]
	b.active = append(b.active, b.writers[0])
	b.selected = b.selected[:0]
	for _, r := range b.results {
		for _, a := range r {
			a.Release()
		}
	}
	b.results = b.results[:0]
}

func (b *StoreBuilder[T]) process(ctx context.Context, query Query) func(io.ReaderAt, int64) error {
	return func(ra io.ReaderAt, i int64) error {
		return b.processFile(ctx, ra, i, query)
	}
}

func (b *StoreBuilder[T]) processFile(ctx context.Context, f io.ReaderAt, size int64, query Query) error {
	file, err := parquet.OpenFile(f, size)
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
	var ls []map[string]parquet.Page
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
			pls := make(map[string]parquet.Page)
			pls[b.active[0].name] = p
			for _, w := range b.active[1:] {
				px, err := cs[w.name].ReadPage()
				if err != nil {
					return err
				}
				pls[w.name] = px
			}
			ls = append(ls, pls)
		}
	}
	for _, pages := range ls {
		b.processPages(ctx, pages, query)
	}
	return nil
}

func (b *StoreBuilder[T]) processPages(ctx context.Context, pages map[string]parquet.Page, query Query) error {
	err := b.filterPages(ctx, pages, query)
	if err != nil {
		if errors.Is(err, ErrSkipPage) {
			return nil
		}
		return err
	}
	return nil
}

func (b *StoreBuilder[T]) filterPages(ctx context.Context, pages map[string]parquet.Page, query Query) error {
	a := make([]compute.Datum, len(b.selected))
	for i, s := range b.selected {
		err := s.WritePage(pages[s.name])
		if err != nil {
			return err
		}
		a[i] = compute.NewDatum(s.build.NewArray())
	}
	defer func() {
		for i := range a {
			a[i].Release()
		}
	}()
	size := pages[b.active[0].name].NumValues()
	ls := make([]bool, size)
	values := make([]parquet.Value, size)
	active := make(map[string]struct{})
	for _, w := range b.active[1:] {
		active[w.name] = struct{}{}
	}

	for _, f := range query.filters {
		if _, ok := active[f.field]; ok {
			if f.h != nil {
				if !f.h(ctx, values, ls, pages[f.field]) {
					return ErrSkipPage
				}
			}
		}
	}

	b.builders.Bool.AppendValues(ls, nil)
	f := compute.NewDatum(b.builders.Bool.NewArray())
	// ls contains booleans indicating which row to choose. This means it  applies the
	// same for all columns.
	for j := range a {
		r, err := compute.Filter(ctx, a[j], f, compute.FilterOptions{})
		if err != nil {
			f.Release()
			log.Get(ctx).Err(err).Msg("failed to apply filter")
			return err
		}
		a[j].Release()
		a[j] = r
	}
	for i := range a {
		b.results[i] = append(b.results[i], a[i].(*compute.ArrayDatum).MakeArray())
	}
	return nil
}

type Record struct {
	Labels  map[string]string `json:"labels,omitempty"`
	Columns []string          `json:"columns,omitempty"`
	Values  []arrow.Array     `json:"values,omitempty"`
}

func (b *StoreBuilder[T]) record(ctx context.Context) (*Record, error) {
	r := &Record{}
	for i, s := range b.selected {
		a, err := array.Concatenate(b.results[i], memory.DefaultAllocator)
		if err != nil {
			return nil, err
		}
		r.Columns = append(r.Columns, s.name)
		r.Values = append(r.Values, a)
	}
	return r, nil
}
