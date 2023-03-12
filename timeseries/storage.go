package timeseries

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"unsafe"

	"github.com/apache/arrow/go/v12/arrow"
	"github.com/apache/arrow/go/v12/arrow/array"
	"github.com/apache/arrow/go/v12/arrow/compute"
	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/apache/arrow/go/v12/arrow/scalar"
	"github.com/segmentio/parquet-go"
)

const (
	SortRowCount = int64(4089)
)

var ErrNoRows = errors.New("no rows")
var ErrSkipPage = errors.New("skip page")

func QueryTable(ctx context.Context, uid uint64, query Query, files ...string) (*Record, error) {
	bob := Bob{db: Get(ctx).db}
	b := entriesBuildPool.Get().(*StoreBuilder)
	table := EVENTS
	defer func() {
		b.reset()
		entriesBuildPool.Put(b)
	}()
	err := bob.Iterate(ctx, table, uid, query.start, query.end, b.Process(ctx, query))
	if err != nil {
		return nil, err
	}
	return b.Result(ctx)
}

var entriesBuildPool = &sync.Pool{
	New: func() any {
		return build()
	},
}

func build() *StoreBuilder {
	b := &StoreBuilder{
		names:   make(map[string]*writer),
		writers: make([]*writer, len(Fields)),
		timestamp: array.NewTimestampBuilder(memory.DefaultAllocator, &arrow.TimestampType{
			Unit: arrow.Nanosecond,
		}),
	}

	for i, f := range Fields {
		b.names[f.Name] = &writer{
			build:   array.NewBuilder(memory.DefaultAllocator, f.Type),
			collect: make([]arrow.Array, 1024),
			index:   i,
			name:    f.Name,
		}
	}
	return b
}

type writer struct {
	build   array.Builder
	collect []arrow.Array
	index   int
	pick    bool
	name    string
	filter  *FILTER
	chunk   parquet.ColumnChunk
	pages   parquet.Pages
}

type int64Buf struct {
	values []int64
}

func (i *int64Buf) release() {
	int64Pool.Put(i)
}

type int32Buf struct {
	values []int32
}

func (i *int32Buf) release() {
	int32Pool.Put(i)
}

type boolBuf struct {
	values []bool
}

func (i *boolBuf) release() {
	boolPool.Put(i)
}

type valuesBuf struct {
	values []parquet.Value
}

func (i *valuesBuf) release() {
	valuesPool.Put(i)
}

var int64Pool = &sync.Pool{
	New: func() any {
		return &int64Buf{values: make([]int64, 4098)}
	},
}

var int32Pool = &sync.Pool{
	New: func() any {
		return &int32Buf{values: make([]int32, 4098)}
	},
}

var boolPool = &sync.Pool{
	New: func() any {
		return &boolBuf{values: make([]bool, 4098)}
	},
}

var valuesPool = &sync.Pool{
	New: func() any {
		return &valuesBuf{values: make([]parquet.Value, 4098)}
	},
}

func (w *writer) write(p parquet.Page) error {
	switch b := w.build.(type) {
	case *array.BooleanBuilder:
		buf := boolPool.Get().(*boolBuf)
		defer buf.release()
		values := buf.values
		r := p.Values().(parquet.BooleanReader)
		var n int
		var err error
		for {
			n, err = r.ReadBooleans(values)
			if err != nil {
				if errors.Is(err, io.EOF) {
					b.AppendValues(values[:n], nil)
					break
				}
			}
			b.AppendValues(values[:n], nil)
		}
	case *array.Int32Builder:
		buf := int32Pool.Get().(*int32Buf)
		defer buf.release()
		values := buf.values
		r := p.Values().(parquet.Int32Reader)
		var n int
		var err error
		for {
			n, err = r.ReadInt32s(values)
			if err != nil {
				if errors.Is(err, io.EOF) {
					b.AppendValues(values[:n], nil)
					break
				}
			}
			b.AppendValues(values[:n], nil)
		}
	case *array.TimestampBuilder:
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
					b.AppendValues(makeTimestamp(values[:n]), nil)
					break
				}
			}
			b.AppendValues(makeTimestamp(values[:n]), nil)
		}
	case *array.DurationBuilder:
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
					b.AppendValues(makeDUration(values[:n]), nil)
					break
				}
			}
			b.AppendValues(makeDUration(values[:n]), nil)
		}
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
			b.AppendValues(values[:n], nil)
		}
	case *array.BinaryDictionaryBuilder:
		buf := valuesPool.Get().(*valuesBuf)
		defer buf.release()
		values := buf.values
		var n int
		var err error
		r := p.Values()
		for {
			n, err = r.ReadValues(values)
			if err != nil {
				if errors.Is(err, io.EOF) {
					for i := 0; i < n; i += 1 {
						err = b.Append(values[i].Bytes())
						if err != nil {
							return err
						}
					}
					break
				}
			}
			for i := 0; i < n; i += 1 {
				err = b.Append(values[i].Bytes())
				if err != nil {
					return err
				}
			}
		}
	case *array.FixedSizeBinaryDictionaryBuilder:
		buf := valuesPool.Get().(*valuesBuf)
		defer buf.release()
		values := buf.values
		var n int
		var err error
		r := p.Values()
		for {
			n, err = r.ReadValues(values)
			if err != nil {
				if errors.Is(err, io.EOF) {
					for i := 0; i < n; i += 1 {
						err = b.Append(values[i].Bytes())
						if err != nil {
							return err
						}
					}
					break
				}
			}
			for i := 0; i < n; i += 1 {
				err = b.Append(values[i].Bytes())
				if err != nil {
					return err
				}
			}
		}

	case *array.Int64DictionaryBuilder:
		buf := int64Pool.Get().(*int64Buf)
		defer buf.release()
		values := buf.values
		var n int
		var err error
		r := p.Values().(parquet.Int64Reader)
		for {
			n, err = r.ReadInt64s(values)
			if err != nil {
				if errors.Is(err, io.EOF) {
					for i := 0; i < n; i += 1 {
						err = b.Append(values[i])
						if err != nil {
							return err
						}
					}
					break
				}
			}
			for i := 0; i < n; i += 1 {
				err = b.Append(values[i])
				if err != nil {
					return err
				}
			}
		}

	default:
	}
	return nil
}

func makeDUration(a []int64) []arrow.Duration {
	return *(*[]arrow.Duration)(unsafe.Pointer(&a))
}

type StoreBuilder struct {
	names     map[string]*writer
	pick      []string
	filters   []FILTER
	writers   []*writer
	active    []*writer
	timestamp *array.TimestampBuilder
}

type releasable interface {
	Release()
}

type releaseList []releasable

func (r releaseList) Release() {
	for _, v := range r {
		v.Release()
	}
}

type releaseFunc func()

func (r releaseFunc) Release() {
	r()
}

func (b *StoreBuilder) reset() {
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

func (b *StoreBuilder) Process(ctx context.Context, query Query) func(io.ReaderAt, int64) error {
	return func(ra io.ReaderAt, i int64) error {
		return b.ProcessFile(ctx, ra, i, query)
	}
}

func (b *StoreBuilder) ProcessFile(ctx context.Context, f io.ReaderAt, size int64, query Query) error {
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
		if w, ok := b.names[n]; ok {
			m[n] = struct{}{}
			w.pick = true
			b.active = append(b.active, w)
			b.pick = append(b.pick, n)
		}
	}
	b.filters = append(b.filters, query.filters.build()...)

	for i := range b.filters {
		p := &b.filters[i]
		_, ok := m[p.Field]
		if !ok {
			w := b.names[p.Field]
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

func (e *writer) read() error {
	p, err := e.pages.ReadPage()
	if err != nil {
		return err
	}
	return e.write(p)
}

func (b *StoreBuilder) rows(ctx context.Context, rg parquet.RowGroup, query Query) error {
	columns := rg.ColumnChunks()

	// Ensure pages are closed regardless of success or failure to execute this
	// function.
	var releasePages releaseList
	defer releasePages.Release()

	// RowGroup level filter
	for _, w := range b.active {
		column := columns[w.index]
		w.chunk = column
		w.pages = column.Pages()
		releasePages = append(releasePages, releaseFunc(func() {
			w.pages.Close()
		}))

		if w.filter == nil {
			continue
		}
		switch w.filter.Op {
		case BLOOM_EQ:
			// We guarantee  that the bloom is in memory. If any filter matches the
			//  bloom filter condition we skip the the entire row group.
			ok, _ := column.BloomFilter().Check(w.filter.Parquet)
			if !ok {
				return nil
			}
		case BLOOM_NE:
			// We guarantee  that the bloom is in memory. If any filter matches the
			//  bloom filter condition we skip  the entire row group.
			ok, _ := column.BloomFilter().Check(w.filter.Parquet)
			if ok {
				return nil
			}
		}
	}

	tsw := b.active[0]
	pages := tsw.pages

	start := query.start.UnixNano()
	end := query.start.UnixNano()

	startDatum := compute.NewDatum(
		scalar.NewTimestampScalar(arrow.Timestamp(start), &arrow.TimestampType{
			Unit: arrow.Nanosecond,
		}),
	)

	endDatum := compute.NewDatum(
		scalar.NewTimestampScalar(arrow.Timestamp(end), &arrow.TimestampType{
			Unit: arrow.Nanosecond,
		}),
	)

	// Cleanup per page resources
	var release releaseList
	defer release.Release()
	var lastIteration bool

	for {
		if len(release) > 0 {
			// free resources from last iteration
			release.Release()
			release = release[:0]
		}
		if lastIteration {
			// we have reached the end of our range within this row group. This will
			// be true if end timestamp is found in one of the pages in this row
			// group.
			return ErrSkip
		}

		tsPage, err := pages.ReadPage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if tsPage.NumRows() == 0 {
			// skip empty pages.
			continue
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

			// By now we have the following facts
			//
			//	[start, end] timestamp within bounds
			//	pages contains rows with data we want
			err = b.readTimestamp(tsPage)
			if err != nil {
				return fmt.Errorf("failed to read timestamp %v", err)
			}

			tsArray := b.timestamp.NewArray()
			release = append(release, tsArray)

			var activeFilter compute.Datum
			{
				// We have already positioned start and end to point to the beginning
				// and end of the day respectively.  Data of interest is now
				// reduced to  start < data < end.

				// select values greater then start timestamp
				value0, err := compute.CallFunction(ctx, "greater", nil,
					&compute.ArrayDatum{
						Value: tsArray.Data(),
					},
					startDatum,
				)
				if err != nil {
					return err
				}
				release = append(release, value0)

				// select values less then end timestamp
				value1, err := compute.CallFunction(ctx, "less", nil,
					&compute.ArrayDatum{
						Value: tsArray.Data(),
					},
					endDatum,
				)
				if err != nil {
					return err
				}
				tsMatch, err := compute.CallFunction(ctx, "and", nil, value0, value1)
				if err != nil {
					return err
				}
				release = append(release, tsMatch)

				activeFilter = tsMatch

				// Timestamp is sorted in ascending order. If the last element of
				// tsMatch is false means this is the last page we are supposed
				// to read.
				a := array.NewBooleanData(tsMatch.(*compute.ArrayDatum).Value)
				lastIteration = !a.Value(a.Len() - 1)
				release = append(release, a)
			}

			for _, e := range b.active {
				// avoid timestamp on this stage. We have already read timestamps
				// above.
				if e.name == "timestamp" {
					continue
				}
				// read both selected and filter fields.
				err := e.read()
				if err != nil {
					return err
				}

				if e.filter != nil {
					a := e.build.NewArray()
					release = append(release, a)
					fa, err := compute.CallFunction(ctx, e.filter.Op.String(), nil,
						compute.NewDatumWithoutOwning(a),
						e.filter.Scalar,
					)

					if err != nil {
						return fmt.Errorf("failed to apply filter on field %q %v", e.name, err)
					}
					release = append(release, fa)
					filterMatch, err := compute.CallFunction(ctx, "and", nil, activeFilter, activeFilter)
					if err != nil {
						return err
					}
					if err != nil {
						return fmt.Errorf("failed to apply filter match  on field %q %v", e.name, err)
					}
					activeFilter = filterMatch
				}
			}

			for _, e := range b.active {
				if e.filter == nil {
					a := e.build.NewArray()
					release = append(release, a)
					selected, err := compute.Filter(
						ctx,
						compute.NewDatumWithoutOwning(a),
						activeFilter,
						compute.FilterOptions{},
					)
					if err != nil {
						return err
					}
					release = append(release, selected)
					e.collect = append(e.collect,
						selected.(*compute.ArrayDatum).MakeArray(),
					)
				}
			}
		}
	}
}

// reads timestamp data into b.timestamp
func (b *StoreBuilder) readTimestamp(p parquet.Page) error {
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
				b.timestamp.AppendValues(makeTimestamp(values[:n]), nil)
				break
			}
			return err
		}
		b.timestamp.AppendValues(makeTimestamp(values[:n]), nil)
	}
	return nil
}

func makeTimestamp(a []int64) []arrow.Timestamp {
	return *(*[]arrow.Timestamp)(unsafe.Pointer(&a))
}

type Record struct {
	Columns []string      `json:"columns,omitempty"`
	Values  []arrow.Array `json:"values,omitempty"`
}

func (b *StoreBuilder) Result(ctx context.Context) (*Record, error) {
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
