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
	"github.com/segmentio/parquet-go"
)

const (
	SortRowCount = int64(4089)
	BUFFER       = 4098
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

	err := bob.Iterate(ctx, table, uid, query.start, query.end, b.Process(ctx, query))
	if err != nil {
		return nil, err
	}
	return b.Result(ctx)
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
		names:   make(map[string]*writer),
		writers: make([]*writer, len(fields)),
		boolean: array.NewBooleanBuilder(memory.DefaultAllocator),
	}

	for i, f := range fields {
		dt, err := ParquetNodeToType(f)
		if err != nil {
			panic(err.Error())
		}
		b.names[f.Name()] = &writer{
			build:   array.NewBuilder(memory.DefaultAllocator, dt),
			collect: make([]arrow.Array, 1024),
			index:   i,
			name:    f.Name(),
		}
	}
	return b
}

type writer struct {
	build      array.Builder
	collect    []arrow.Array
	index      int
	pick       bool
	name       string
	filter     *FILTER
	chunk      parquet.ColumnChunk
	dictionary bool
	pages      parquet.Pages
	page       parquet.Page
}

type int64Buf struct {
	values []int64
}

func (i *int64Buf) release() {
	int64Pool.Put(i)
}

type valuesBuf struct {
	values []parquet.Value
}

func (i *valuesBuf) release() {
	valuesPool.Put(i)
}

type stringsBuf struct {
	values []string
}

func (i *stringsBuf) release() {
	stringsPool.Put(i)
}

type boolBuf struct {
	values []bool
}

func (i *boolBuf) release() {
	i.values = i.values[:0]
	stringsPool.Put(i)
}

func (i *boolBuf) reserve(size int) []bool {
	i.values = i.values[:0]
	n := cap(i.values)
	if size <= cap(i.values) {
		i.values = i.values[:size]
	} else {
		i.values = make([]bool, n*size)
	}
	for x := 0; x < size; x += 1 {
		i.values[x] = false
	}
	return i.values
}

var int64Pool = &sync.Pool{
	New: func() any {
		return &int64Buf{values: make([]int64, 4098)}
	},
}

var valuesPool = &sync.Pool{
	New: func() any {
		return &valuesBuf{values: make([]parquet.Value, 4098)}
	},
}

var stringsPool = &sync.Pool{
	New: func() any {
		return &stringsBuf{values: make([]string, 4098)}
	},
}

var boolBufPool = &sync.Pool{
	New: func() any {
		return &boolBuf{values: make([]bool, 4098)}
	},
}

func (w *writer) write(p parquet.Page, filter []bool, f func(any) bool) error {
	switch b := w.build.(type) {
	case *array.Int64Builder:
		buf := int64Pool.Get().(*int64Buf)
		defer buf.release()
		values := buf.values
		r := p.Values().(parquet.Int64Reader)
		var n int
		var err error
		for {
			n, err = r.ReadInt64s(values)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
			}
			if f != nil {
				for i := 0; i < n; i += 1 {
					if f(values[i]) {
						filter[i] = true
					}
				}
			}

			filter = filter[n:]
			if w.pick {
				b.AppendValues(values[:n], nil)
			}
		}
	case *array.StringBuilder:
		buf := valuesPool.Get().(*valuesBuf)
		defer buf.release()
		values := buf.values
		sb := stringsPool.Get().(*stringsBuf)
		sv := sb.values
		defer sb.release()
		var n int
		var err error
		r := p.Values()
		for {
			n, err = r.ReadValues(values)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
			}
			if f != nil {
				for i := 0; i < n; i++ {
					sv[i] = values[i].String()
					if f(sv[i]) {
						filter[i] = true
					}
				}
			}

			filter = filter[n:]
			if w.pick {
				b.AppendValues(sv[:n], nil)
			}
		}
	default:
	}
	return nil
}

type StoreBuilder[T any] struct {
	names   map[string]*writer
	pick    []FIELD_TYPE
	filters []FILTER
	writers []*writer
	active  []*writer
	boolean *array.BooleanBuilder
}

func (b *StoreBuilder[T]) reset() {
	b.pick = b.pick[:0]
	b.filters = b.filters[:0]
	b.active = b.active[:0]
	for _, w := range b.writers {
		for i := 0; i < len(w.collect); i += 1 {
			w.collect[i].Release()
		}
		w.collect = w.collect[:0]
	}
}

func (b *StoreBuilder[T]) Process(ctx context.Context, query Query) func(io.ReaderAt, int64) error {
	return func(ra io.ReaderAt, i int64) error {
		return b.ProcessFile(ctx, ra, i, query)
	}
}

func (b *StoreBuilder[T]) ProcessFile(ctx context.Context, f io.ReaderAt, size int64, query Query) error {
	file, err := parquet.OpenFile(f, size)
	if err != nil {
		return err
	}
	// always select timestamp
	ts := b.names["timestamp"]
	ts.pick = true
	b.active = append(b.active, ts)
	m := make(map[string]struct{})

	for _, n := range query.selected {
		f := TypeFromString(n)
		if f == UNKNOWN {
			continue
		}
		m[f.String()] = struct{}{}
		w := b.names[f.String()]
		w.pick = true
		b.active = append(b.active, w)
		b.pick = append(b.pick, f)
	}
	b.filters = append(b.filters, query.filters.build()...)

	for i := range b.filters {
		p := &b.filters[i]
		_, ok := m[p.Field.Type().String()]
		if !ok {
			w := b.names[p.Field.Type().String()]
			w.filter = p
			b.active = append(b.active, w)
		}
	}

	for _, rg := range file.RowGroups() {
		err = b.rows(ctx, rg, query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *writer) skip() (bool, error) {
	if e.dictionary {
		p, err := e.pages.ReadPage()
		if err != nil {
			return false, err
		}
		e.page = p
		value := e.filter.Field.Value()
		dict := p.Dictionary()
		for i := 0; i < dict.Len(); i += 1 {
			if parquet.Equal(dict.Index(int32(i)), value) {
				return false, nil
			}
		}
		// no hash with the value on the this page.
		return true, nil
	}
	return false, nil
}

func (e *writer) read(b []bool) error {
	if e.page != nil {
		// we performed a dict filter before the page is already open we just read it here
		// and set it to nil so next reads will open a new page.
		err := e.write(e.page, b, e.match())
		e.page = nil
		return err
	}
	p, err := e.pages.ReadPage()
	if err != nil {
		return err
	}
	return e.write(p, b, e.match())
}

func (e *writer) final(ctx context.Context, filter arrow.Array) error {
	if e.pick {
		a, err := compute.FilterArray(ctx, e.build.NewArray(), filter, *compute.DefaultFilterOptions())
		if err != nil {
			return err
		}
		e.collect = append(e.collect, a)
	}
	return nil
}

func (e *writer) match() func(any) bool {
	if e.filter == nil {
		return nil
	}
	return e.filter.Match
}

func (b *StoreBuilder[T]) rows(ctx context.Context, rg parquet.RowGroup, query Query) error {
	columns := rg.ColumnChunks()
	// map all columns to fields
	fields := make(map[FIELD_TYPE]*writer)

	// RowGroup level filter
	for _, w := range b.active {
		column := columns[w.index]
		w.chunk = column
		w.pages = column.Pages()
		if w.filter == nil {
			// set column and
			continue
		}
		w.dictionary = w.filter.Op == BLOOM_AND_DICT_EQ ||
			w.filter.Op == DICT
		switch w.filter.Op {
		case BLOOM_EQ, BLOOM_AND_DICT_EQ:
			// We guarantee  that the bloom is in memory. If any filter matches the
			//  bloom filter condition we skip the the entire row group.
			ok, _ := column.BloomFilter().Check(w.filter.Field.Value())
			if !ok {
				return nil
			}
		case BLOOM_NE:
			// We guarantee  that the bloom is in memory. If any filter matches the
			//  bloom filter condition we skip the the entire row group.
			ok, _ := column.BloomFilter().Check(w.filter.Field.Value())
			if ok {
				return nil
			}
		}
	}

	// everything is relative to the timestamp. Timestamps are stored as unix nanoseconds
	tsw := b.active[0]
	ts := columns[tsw.index]
	pages := ts.Pages()
	defer pages.Close()

	start := query.start.UnixNano()
	end := query.start.UnixNano()

	f := func(a any) bool {
		x := a.(int64)
		return start <= x && x < end
	}

	xb := boolBufPool.Get().(*boolBuf)
	defer xb.release()

	for {
		tsPage, err := pages.ReadPage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		min, max, ok := tsPage.Bounds()
		if !ok {
			continue
		}

		minValue := min.Int64()
		maxValue := max.Int64()
		if minValue >= end {
			// NOTE: we need to signal the end of file processing. Any page encountered
			// whose minimum value(timestamp) exceeds the last timestamp we want in our
			// range query denotes the end of our search.
			return ErrSkip
		}

		if minValue < end && maxValue > start {
			// The page is within bounds

			// first we quickly use dictionary to skip pages not containing fields
			// with dict type.
			var skip bool
			for _, f := range fields {
				skip, err = f.skip()
				if err != nil {
					return err
				}
				if skip {
					// no need to process any other filter
					break
				}
			}
			if skip {
				// skip this page, move on to the next one.
				continue
			}

			// By now we have the following facts
			//
			//	[start, end] timestamp within bounds
			//	pages contains rows with data we want
			//

			filter := xb.reserve(int(tsPage.NumRows()))

			// directly read the timestamp data
			err = tsw.write(tsPage, filter, f)
			if err != nil {
				return err
			}

			//  select everything
			for _, e := range fields {
				// avoid timestamp on this stage. We have already read timestamps
				// above.
				if e.name == "timestamp" {
					continue
				}
				err := e.read(filter)
				if err != nil {
					return err
				}
			}

			// we have applied initial filters based on parquet. Now we use arrow to
			// take picked rows.
			//
			// Same filter is applied to all selected columns
			b.boolean.AppendValues(filter, nil)
			a := b.boolean.NewArray()
			// e.collect will only work with selected writers so, just range over the fields
			// without worrying whether they are filter fields or selected fields.
			for _, e := range fields {
				err := e.final(ctx, a)
				if err != nil {
					a.Release()
					return err
				}
			}
			a.Release()
		}
	}

}

type Record struct {
	Columns []string      `json:"columns,omitempty"`
	Values  []arrow.Array `json:"values,omitempty"`
}

func (b *StoreBuilder[T]) Result(ctx context.Context) (*Record, error) {
	r := &Record{}
	for _, w := range b.writers {
		if w.pick {
			a, err := array.Concatenate(w.collect, memory.DefaultAllocator)
			if err != nil {
				return nil, err
			}
			r.Columns = append(r.Columns, w.name)
			r.Values = append(r.Values, a)
		}
	}
	return r, nil
}
