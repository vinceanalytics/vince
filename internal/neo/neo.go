package neo

import (
	"context"
	"errors"
	"io"
	"path"
	reflect "reflect"
	"regexp"
	"time"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/compute"
	"github.com/apache/arrow/go/v13/arrow/memory"
	"github.com/parquet-go/parquet-go"
	"github.com/vinceanalytics/vince/pkg/log"
	"golang.org/x/exp/slices"
)

type FilterType uint

const (
	EXACT FilterType = iota
	GLOB
	RE
)

type IFilter interface {
	Field() string
	Accept(g parquet.RowGroup) bool
	Pages(g parquet.RowGroup) IPageReader
	Release()
}

type IFilterList []IFilter

func (ls IFilterList) Accept(g parquet.RowGroup) bool {
	for _, f := range ls {
		if !f.Accept(g) {
			return false
		}
	}
	return true
}

type IPageReader interface {
	Page(*Booleans, ...func(arrow.Array)) error
	Close() error
}

type FieldReader interface {
	Pages(g parquet.RowGroup) IPageReader
}

type KeyValue[T any] struct {
	Key   string
	Op    FilterType
	Value T
}

func StringBloomFilter(value string, op FilterType, names ...string) func(g parquet.RowGroup) bool {
	if op != EXACT {
		return func(g parquet.RowGroup) bool {
			return true
		}
	}
	pv := parquet.ValueOf(value)
	return func(g parquet.RowGroup) bool {
		schema := g.Schema()
		leaf, ok := schema.Lookup(names...)
		if !ok {
			return ok
		}
		col := g.ColumnChunks()[leaf.ColumnIndex]
		if b := col.BloomFilter(); b != nil {
			ok, err := b.Check(pv)
			if err != nil {
				log.Get().Fatal().Err(err).Msg("failed reading bloom filter")
			}
			return ok
		}
		// If the string field does not have a bloom filter always select the row
		// group for processing.
		return true
	}
}

func StringMatch(val string, op FilterType) func(parquet.Value) bool {
	if op == EXACT {
		value := parquet.ValueOf(val)
		return func(v parquet.Value) bool {
			return parquet.Equal(v, value)
		}
	}
	if op == GLOB {
		return func(v parquet.Value) bool {
			ok, _ := path.Match(val, v.String())
			return ok
		}
	}
	re := regexp.MustCompile(val)
	return func(v parquet.Value) bool {
		return re.MatchString(v.String())
	}
}

func FilterString(o KeyValue[string], names ...string) FilterFuncs {
	values := make([]parquet.Value, 0, 1<<10)
	match := StringMatch(o.Value, o.Op)
	build := array.NewStringBuilder(memory.DefaultAllocator)
	return FilterFuncs{
		ReleaseFunc: build.Release,
		AcceptFunc:  StringBloomFilter(o.Value, o.Op, names...),
		PagesFunc: func(g parquet.RowGroup) IPageReader {
			schema := g.Schema()
			leaf, ok := schema.Lookup(names...)
			if !ok {
				// If the row group does not contain the field we are supposed to filter.
				// Don't select anything from this row group.
				return nil
			}
			col := g.ColumnChunks()[leaf.ColumnIndex]
			pages := col.Pages()
			return PageReaderFuncs{
				CloseFunc: pages.Close,
				PageFunc: func(fb *Booleans, fn ...func(arrow.Array)) error {
					b := fb.Get()
					p, err := pages.ReadPage()
					if err != nil {
						if !errors.Is(err, io.EOF) {
							log.Get().Fatal().Err(err).Msg("failed to read a page")
						}
						return err
					}
					n := p.NumValues()
					values = slices.Grow(values, int(n))
					_, err = p.Values().ReadValues(values)
					if err != nil {
						if !errors.Is(err, io.EOF) {
							log.Get().Fatal().Err(err).Msg("failed to read a page")
						}
					}
					pick := len(fn) > 0
					for i := range values {
						b[i] = b[i] && match(values[i])
						if pick {
							build.Append(values[i].String())
						}
					}
					if pick {
						fn[0](build.NewArray())
					}
					return nil
				},
			}
		},
	}
}

type ValueMatch struct {
	Op    FilterType
	Value string
}

type MapMatch struct {
	Field string
	Match map[string]ValueMatch
}

func FilterMap(o MapMatch) FilterFuncs {
	build := array.NewMapBuilder(memory.DefaultAllocator,
		&arrow.StringType{},
		&arrow.StringType{}, false,
	)
	keyPath := []string{o.Field, "Key_value", "key"}
	valuePath := []string{o.Field, "Key_value", "value"}

	var bloomKeys []parquet.Value
	var bloomValues []parquet.Value

	var keyMatch []string
	var valueMatch []func(parquet.Value) bool

	for k, v := range o.Match {
		bloomKeys = append(bloomKeys, parquet.ValueOf(k))
		if v.Op == EXACT {
			bloomValues = append(bloomValues, parquet.ValueOf(v.Value))
		}
		keyMatch = append(keyMatch, k)
		valueMatch = append(valueMatch, StringMatch(v.Value, v.Op))
	}
	keys := make([]parquet.Value, 0, 1<<10)
	values := make([]parquet.Value, 0, 1<<10)

	keyBuild := build.KeyBuilder().(*array.StringBuilder)
	valueBuild := build.ValueBuilder().(*array.StringBuilder)

	match := func(m map[string]parquet.Value) bool {
		for i := range keyMatch {
			v, ok := m[keyMatch[i]]
			if !ok {
				return false
			}
			if !valueMatch[i](v) {
				return false
			}
		}
		return true
	}
	return FilterFuncs{
		ReleaseFunc: build.Release,
		AcceptFunc: func(g parquet.RowGroup) bool {
			scheme := g.Schema()
			cols := g.ColumnChunks()
			{
				// check for keys
				l, ok := scheme.Lookup(keyPath...)
				if !ok {
					return ok
				}
				b := cols[l.ColumnIndex].BloomFilter()
				if b == nil {
					return false
				}
				for _, k := range bloomKeys {
					ok, err := b.Check(k)
					if err != nil {
						log.Get().Fatal().Err(err).Msg("failed to check value from bloom filter")
					}
					if !ok {
						return ok
					}
				}

			}
			{
				// check for values
				l, ok := scheme.Lookup(valuePath...)
				if !ok {
					return ok
				}
				b := cols[l.ColumnIndex].BloomFilter()
				if b == nil {
					return false
				}
				for _, k := range bloomValues {
					ok, err := b.Check(k)
					if err != nil {
						log.Get().Fatal().Err(err).Msg("failed to check value from bloom filter")
					}
					if !ok {
						return ok
					}
				}
			}
			return true
		},
		PagesFunc: func(g parquet.RowGroup) IPageReader {
			schema := g.Schema()
			var keysPages, valuesPages parquet.Pages
			{
				// read keys
				leaf, ok := schema.Lookup(keyPath...)
				if !ok {
					// If the row group does not contain the field we are supposed to filter.
					// Don't select anything from this row group.
					return nil
				}
				col := g.ColumnChunks()[leaf.ColumnIndex]
				keysPages = col.Pages()
			}
			{
				// read values
				leaf, ok := schema.Lookup(valuePath...)
				if !ok {
					// If the row group does not contain the field we are supposed to filter.
					// Don't select anything from this row group.
					return nil
				}
				col := g.ColumnChunks()[leaf.ColumnIndex]
				valuesPages = col.Pages()
			}

			return PageReaderFuncs{
				CloseFunc: func() error {
					return errors.Join(keysPages.Close(), valuesPages.Close())
				},
				PageFunc: func(fb *Booleans, fn ...func(arrow.Array)) error {
					b := fb.Get()
					{
						// read keys
						p, err := keysPages.ReadPage()
						if err != nil {
							return err
						}
						n := p.NumValues()
						keys = slices.Grow(keys, int(n))[:n]
						_, err = p.Values().ReadValues(values)
						if err != nil {
							if !errors.Is(err, io.EOF) {
								log.Get().Fatal().Err(err).Msg("failed to read a page")
							}
						}
					}
					{
						// read values
						p, err := valuesPages.ReadPage()
						if err != nil {
							return err
						}
						n := p.NumValues()
						values = slices.Grow(values, int(n))[:n]
						_, err = p.Values().ReadValues(values)
						if err != nil {
							if !errors.Is(err, io.EOF) {
								log.Get().Fatal().Err(err).Msg("failed to read a page")
							}
						}
					}

					pick := len(fn) > 0

					m := make(map[string]parquet.Value)
					var bIdx int
					for i := range keys {
						x := keys[i]
						if x.RepetitionLevel() == 0 {
							if i != 0 {
								// We have collected a key/value pairs in this  row
								if pick {
									build.Append(true)
									for k, v := range m {
										keyBuild.Append(k)
										valueBuild.Append(v.String())
									}
								}
								b[bIdx] = b[bIdx] && match(m)
								for k := range m {
									delete(m, k)
								}
								bIdx++
							}
						}
					}
					b[bIdx] = b[bIdx] && match(m)
					if pick {
						build.Append(true)
						for k, v := range m {
							keyBuild.Append(k)
							valueBuild.Append(v.String())
						}
						fn[0](build.NewArray())
					}
					return nil
				},
			}
		},
	}
}

var ErrSkipPage = errors.New("skip page")
var ErrComplete = errors.New("complete page read")

func FilterTimestamp(start, end time.Time) FilterFuncs {
	build := array.NewTimestampBuilder(memory.DefaultAllocator, &arrow.TimestampType{
		Unit:     arrow.Millisecond,
		TimeZone: "UTC",
	})
	values := make([]int64, 0, 1<<10)
	lo := start.UnixMilli()
	hi := end.UnixMilli()
	return FilterFuncs{
		FieldName:   "timestamp",
		ReleaseFunc: build.Release,
		AcceptFunc: func(g parquet.RowGroup) bool {
			// We don't use bloom filters on timestamp field.
			return true
		},
		PagesFunc: func(g parquet.RowGroup) IPageReader {
			schema := g.Schema()
			leaf, ok := schema.Lookup("timestamp")
			if !ok {
				return nil
			}
			pages := g.ColumnChunks()[leaf.ColumnIndex].Pages()
			return PageReaderFuncs{
				CloseFunc: pages.Close,
				PageFunc: func(fb *Booleans, f ...func(arrow.Array)) error {
					p, err := pages.ReadPage()
					if err != nil {
						return err
					}
					min, max, ok := p.Bounds()
					if !ok {
						return ErrSkipPage
					}
					lowerBound := min.Int64()
					upperBound := max.Int64()
					if upperBound < lo {
						// Ww can't observe starting timestamp on this page. Move on to the next
						// one
						return ErrSkipPage
					}
					if lowerBound > hi {
						// This page starts way past the end of the query range. Stop Immediately
						return io.EOF
					}
					n := p.NumValues()
					b := fb.Reserve(int(n))
					values = slices.Grow(values, int(n))[:n]
					_, err = p.Values().(parquet.Int64Reader).ReadInt64s(values)
					if err != nil {
						if !errors.Is(err, io.EOF) {
							return err
						}
					}
					for i := 0; i < len(values); i++ {
						if values[i] <= lo {
							b[i] = false
							continue
						}
						break
					}
					for i := len(values) - 1; i > 0; i-- {
						if values[i] >= hi {
							b[i] = false
							continue
						}
						break
					}
					if len(f) > 0 {
						build.Reserve(len(values))
						for i := range values {
							build.Append(arrow.Timestamp(values[i]))
						}
						f[0](build.NewArray())
					}
					if hi < upperBound {
						return ErrComplete
					}
					return nil
				},
			}
		},
	}
}

type FilterFuncs struct {
	FieldName   string
	AcceptFunc  func(g parquet.RowGroup) bool
	PagesFunc   func(g parquet.RowGroup) IPageReader
	ReleaseFunc func()
}

var _ IFilter = (*FilterFuncs)(nil)

func (f FilterFuncs) Field() string {
	return f.FieldName
}

func (f FilterFuncs) Accept(g parquet.RowGroup) bool {
	if f.AcceptFunc != nil {
		return f.AcceptFunc(g)
	}
	return false
}

func (f FilterFuncs) Pages(g parquet.RowGroup) IPageReader {
	if f.PagesFunc != nil {
		return f.PagesFunc(g)
	}
	return noopFilterPage{}
}

func (f FilterFuncs) Release() {
	if f.ReleaseFunc != nil {
		f.ReleaseFunc()
	}
}

type FilterFuncsWithCallback struct {
	IFilter
	Callback func(arrow.Array)
}

func (f FilterFuncsWithCallback) Pages(g parquet.RowGroup) IPageReader {
	o := f.IFilter.Pages(g)
	return PageReaderFuncs{
		Callback:  f.Callback,
		PageFunc:  o.Page,
		CloseFunc: o.Close,
	}
}

type PageReaderFuncs struct {
	Callback  func(arrow.Array)
	PageFunc  func(*Booleans, ...func(arrow.Array)) error
	CloseFunc func() error
}

var _ IPageReader = (*PageReaderFuncs)(nil)

func (f PageReaderFuncs) Page(ok *Booleans, fn ...func(arrow.Array)) error {
	if f.PageFunc != nil {
		if f.Callback != nil {
			return f.PageFunc(ok, f.Callback)
		}
		return f.PageFunc(ok, fn...)
	}
	return nil
}
func (f PageReaderFuncs) Close() error {
	if f.CloseFunc != nil {
		return f.CloseFunc()
	}
	return nil
}

type noopFilterPage struct{}

var _ IPageReader = (*noopFilterPage)(nil)

func (noopFilterPage) Page(_ *Booleans, _ ...func(arrow.Array)) error {
	return nil
}
func (noopFilterPage) Close() error {
	return nil
}

type Options struct {
	Filters    IFilterList
	Start, End time.Time
	Select     []string
}

func (o Options) Entry() Options {
	return Options{
		Filters: o.Filters,
		Start:   o.Start,
		End:     o.End,
		Select: append(o.Select,
			"id",
			"bounce",
			"value",
		),
	}
}

type Booleans struct {
	data []bool
}

func (b *Booleans) Reserve(n int) []bool {
	b.data = slices.Grow(b.data, n)[:n]
	copyBool(b.data)
	return b.data
}

func (b *Booleans) Get() []bool {
	return b.data
}

type GroupProcess func(g parquet.RowGroup) error

func Exec[T any](ctx context.Context, o Options, source func(GroupProcess) error) (arrow.Record, error) {
	selected := buildSelection[T](o.Select...)

	filters := slices.Clone(o.Filters)
	filters = append(IFilterList{FilterTimestamp(o.Start, o.End)}, filters...)

	for i, x := range filters {
		if f, ok := selected[x.Field()]; ok {
			// We are performing filter on a selected field.We wrap the filter in a
			// callback to store the read column page.
			filters[i] = &FilterFuncsWithCallback{
				IFilter:  filters[i],
				Callback: f.SetPage,
			}
			f.Filtered = true
		}
	}

	booleans := &Booleans{
		data: make([]bool, 0, 1<<10),
	}
	pages := make([]IPageReader, 0, len(filters))
	build := array.NewBooleanBuilder(memory.DefaultAllocator)
	defer build.Release()

	err := source(func(g parquet.RowGroup) error {
		if !filters.Accept(g) {
			return nil
		}
		pages = pages[:0]
		for _, f := range filters {
			pgs := f.Pages(g)
			if pgs == nil {
				// Move to the next row group if any of the filters rejects this row group.
				return nil
			}
			pages = append(pages, pgs)
		}
		// we add the selection pages at the end of the filter chain
		for _, f := range selected {
			if f.Filtered {
				// This field pages are already added.
				continue
			}
			pages = append(pages, f.Reader.Pages(g))
		}
		var complete bool
	group:
		for {
			for i, p := range pages {
				err := p.Page(booleans)
				if err != nil {
					if errors.Is(err, io.EOF) {
						if i != 0 {
							log.Get().Fatal().Msg("observed io.EOF error on non timestamp filter")
						}
						// reached the end of this row group. io.EOF is returned before any other
						// filters are applied.
						//
						// This marks the end of processing this row group. We may however still
						// process more row groups
						break group
					}
					if errors.Is(err, ErrSkipPage) {
						if i != 0 {
							log.Get().Fatal().Msg("observed ErrSkipPage error on non timestamp filter")
						}
						continue
					}
					if !errors.Is(err, ErrComplete) {
						return err
					}

					if i != 0 {
						log.Get().Fatal().Msg("observed ErrSkipPage error on non timestamp filter")
					}
					// We have completed processing All groups. There is still data that we
					// must read from this page. So we set complete = true and carry on until
					// the end of this page processing where we explicitly state to not need
					// any more row groups.
					complete = true
				}
			}
			build.AppendValues(booleans.Get(), nil)
			a := build.NewArray()
			for i := range selected {
				selected[i].Apply(ctx, a)
			}
			a.Release()
		}
		for _, p := range pages {
			p.Close()
		}
		if complete {
			return io.EOF
		}
		return nil
	})
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}
	fields := make([]arrow.Field, 0, len(selected))
	values := make([]arrow.Array, 0, len(selected))
	for _, f := range selected {
		fields = append(fields, arrow.Field{
			Name:     f.Name,
			Type:     f.Builder.Type(),
			Nullable: true,
		})
		v, err := array.Concatenate(f.Table, memory.DefaultAllocator)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return array.NewRecord(arrow.NewSchema(fields, nil),
		values, int64(values[0].Len()),
	), nil
}

type Field struct {
	Name     string
	Filtered bool
	Schema   *parquet.Schema
	Reader   FieldReader
	Builder  array.Builder
	Page     arrow.Array
	Table    []arrow.Array
}

var _ FieldReader = (*FilterFuncs)(nil)

func (f *Field) SetPage(a arrow.Array) {
	f.Page = a
}

func (f *Field) Apply(ctx context.Context, ok arrow.Array) {
	o := compute.FilterOptions{}
	r, err := compute.FilterArray(ctx, f.Page, ok, o)
	if err != nil {
		log.Get().Fatal().Err(err).Msg("failed to apply filter")
	}
	f.Table = append(f.Table, r)
	f.Page.Release()
	f.Page = nil
}

var boolBuffer = func() (o []bool) {
	o = make([]bool, 5<<10)
	for i := range o {
		o[i] = true
	}
	return
}()

func copyBool(o []bool) {
	if len(o) < len(boolBuffer) {
		copy(o, boolBuffer[:len(o)])
		return
	}
	for i := range o {
		o[i] = true
	}
}

func buildSelection[T any](roots ...string) (o map[string]*Field) {
	var t T
	schema := parquet.SchemaOf(t)
	o = make(map[string]*Field)
	for _, r := range roots {
		o[r] = nil
	}
	for _, f := range schema.Fields() {
		if _, ok := o[f.Name()]; !ok {
			continue
		}
		field := &Field{
			Name: f.Name(),
		}
		switch e := reflect.Zero(f.GoType()).Interface().(type) {
		case uint64, int64:
			field.Builder = array.NewInt64Builder(memory.DefaultAllocator)
			field.Reader = readInt64(field)
		case time.Duration:
			field.Builder = array.NewDurationBuilder(memory.DefaultAllocator,
				&arrow.DurationType{
					Unit: arrow.Second,
				})
			field.Reader = readDuration(field)
		case time.Time:
			field.Builder = array.NewTimestampBuilder(memory.DefaultAllocator,
				&arrow.TimestampType{
					Unit:     arrow.Millisecond,
					TimeZone: "UTC",
				})
			field.Reader = readTime(field)
		case map[string]string:
			field.Builder = array.NewMapBuilder(
				memory.DefaultAllocator,
				&arrow.StringType{},
				&arrow.StringType{},
				false,
			)
			field.Reader = readMap(field)
		case string:
			field.Builder = array.NewStringBuilder(memory.DefaultAllocator)
			field.Reader = readString(field)
		case float64:
			field.Builder = array.NewFloat64Builder(memory.DefaultAllocator)
			field.Reader = readFloat64(field)
		default:
			log.Get().Fatal().Msgf("unsupported field type %#T", e)
		}
		o[f.Name()] = field
	}
	return
}

func readTime(f *Field) FilterFuncs {
	b := array.NewTimestampBuilder(memory.DefaultAllocator, &arrow.TimestampType{
		Unit: arrow.Millisecond,
	})
	values := make([]int64, 0, 1<<10)
	o := make([]arrow.Timestamp, 0, 1<<10)
	return FilterFuncs{
		ReleaseFunc: b.Release,
		PagesFunc: func(g parquet.RowGroup) IPageReader {
			l, ok := g.Schema().Lookup(f.Name)
			if !ok {
				return nil
			}
			pages := g.ColumnChunks()[l.ColumnIndex].Pages()
			return PageReaderFuncs{
				CloseFunc: pages.Close,
				PageFunc: func(xx *Booleans, _ ...func(arrow.Array)) error {
					fb := xx.Get()
					p, err := pages.ReadPage()
					if err != nil {
						return err
					}
					n := p.NumValues()
					values = slices.Grow(values, int(n))[:n]
					_, err = p.Values().(parquet.Int64Reader).ReadInt64s(values)
					if err != nil {
						if !errors.Is(err, io.EOF) {
							return err
						}
					}
					o = slices.Grow(o, int(n))[:n]
					for i := range values {
						o[i] = arrow.Timestamp(values[i])
					}
					b.AppendValues(o, fb)
					f.SetPage(b.NewArray())
					return nil
				},
			}
		},
	}
}
func readDuration(f *Field) FilterFuncs {
	b := array.NewDurationBuilder(memory.DefaultAllocator, &arrow.DurationType{})
	values := make([]int64, 0, 1<<10)
	o := make([]arrow.Duration, 0, 1<<10)
	return FilterFuncs{
		ReleaseFunc: b.Release,
		PagesFunc: func(g parquet.RowGroup) IPageReader {
			l, ok := g.Schema().Lookup(f.Name)
			if !ok {
				return nil
			}
			pages := g.ColumnChunks()[l.ColumnIndex].Pages()
			return PageReaderFuncs{
				CloseFunc: pages.Close,
				PageFunc: func(fb *Booleans, _ ...func(arrow.Array)) error {
					p, err := pages.ReadPage()
					if err != nil {
						return err
					}
					n := p.NumValues()
					values = slices.Grow(values, int(n))[:n]
					_, err = p.Values().(parquet.Int64Reader).ReadInt64s(values)
					if err != nil {
						if !errors.Is(err, io.EOF) {
							return err
						}
					}
					o = slices.Grow(o, int(n))[:n]
					for i := range values {
						o[i] = arrow.Duration(values[i])
					}
					b.AppendValues(o, fb.Get())
					f.SetPage(b.NewArray())
					return nil
				},
			}
		},
	}
}
func readFloat64(f *Field) FilterFuncs {
	b := array.NewFloat64Builder(memory.DefaultAllocator)
	values := make([]float64, 0, 1<<10)
	return FilterFuncs{
		ReleaseFunc: b.Release,
		PagesFunc: func(g parquet.RowGroup) IPageReader {
			l, ok := g.Schema().Lookup(f.Name)
			if !ok {
				return nil
			}
			pages := g.ColumnChunks()[l.ColumnIndex].Pages()
			return PageReaderFuncs{
				CloseFunc: pages.Close,
				PageFunc: func(fb *Booleans, _ ...func(arrow.Array)) error {
					p, err := pages.ReadPage()
					if err != nil {
						return err
					}
					n := p.NumValues()
					values = slices.Grow(values, int(n))[:n]
					_, err = p.Values().(parquet.DoubleReader).ReadDoubles(values)
					if err != nil {
						if !errors.Is(err, io.EOF) {
							return err
						}
					}
					b.AppendValues(values, fb.Get())
					f.SetPage(b.NewArray())
					return nil
				},
			}
		},
	}
}
func readInt64(f *Field) FilterFuncs {
	b := array.NewInt64Builder(memory.DefaultAllocator)
	values := make([]int64, 0, 1<<10)
	return FilterFuncs{
		ReleaseFunc: b.Release,
		PagesFunc: func(g parquet.RowGroup) IPageReader {
			l, ok := g.Schema().Lookup(f.Name)
			if !ok {
				return nil
			}
			pages := g.ColumnChunks()[l.ColumnIndex].Pages()
			return PageReaderFuncs{
				CloseFunc: pages.Close,
				PageFunc: func(fb *Booleans, _ ...func(arrow.Array)) error {
					p, err := pages.ReadPage()
					if err != nil {
						return err
					}
					n := p.NumValues()
					values = slices.Grow(values, int(n))[:n]
					_, err = p.Values().(parquet.Int64Reader).ReadInt64s(values)
					if err != nil {
						if !errors.Is(err, io.EOF) {
							return err
						}
					}
					b.AppendValues(values, fb.Get())
					f.SetPage(b.NewArray())
					return nil
				},
			}
		},
	}
}

func readString(f *Field) FilterFuncs {
	b := array.NewStringBuilder(memory.DefaultAllocator)
	values := make([]parquet.Value, 0, 1<<10)
	o := make([]string, 0, 1<<10)
	return FilterFuncs{
		ReleaseFunc: b.Release,
		PagesFunc: func(g parquet.RowGroup) IPageReader {
			l, ok := g.Schema().Lookup(f.Name)
			if !ok {
				return nil
			}
			pages := g.ColumnChunks()[l.ColumnIndex].Pages()
			return PageReaderFuncs{
				CloseFunc: pages.Close,
				PageFunc: func(fb *Booleans, _ ...func(arrow.Array)) error {
					p, err := pages.ReadPage()
					if err != nil {
						return err
					}
					n := p.NumValues()
					values = slices.Grow(values, int(n))[:n]
					_, err = p.Values().ReadValues(values)
					if err != nil {
						if !errors.Is(err, io.EOF) {
							return err
						}
					}
					o := slices.Grow(o, int(n))[:n]
					for i := range values {
						o[i] = values[i].String()
					}
					b.AppendStringValues(o, fb.Get())
					f.SetPage(b.NewArray())
					return nil
				},
			}
		},
	}
}
func readMap(f *Field) FilterFuncs {
	b := array.NewMapBuilder(memory.DefaultAllocator,
		&arrow.StringType{},
		&arrow.StringType{},
		false,
	)
	bk := b.KeyBuilder().(*array.StringBuilder)
	bv := b.KeyBuilder().(*array.StringBuilder)
	return FilterFuncs{
		PagesFunc: func(g parquet.RowGroup) IPageReader {
			schema := g.Schema()
			k, ok := schema.Lookup(f.Name, "key_value", "key")
			if !ok {
				return nil
			}
			v, ok := schema.Lookup(f.Name, "key_value", "value")
			if !ok {
				return nil
			}
			cols := g.ColumnChunks()
			keyPages := cols[k.ColumnIndex].Pages()
			valuePages := cols[v.ColumnIndex].Pages()
			keys := make([]parquet.Value, 0, 1<<10)
			values := make([]parquet.Value, 0, 1<<10)

			var ks, vs []string
			return PageReaderFuncs{
				CloseFunc: func() error {
					return errors.Join(keyPages.Close(), valuePages.Close())
				},
				PageFunc: func(fb *Booleans, f ...func(arrow.Array)) error {
					{
						p, err := keyPages.ReadPage()
						if err != nil {
							return err
						}
						n := p.NumValues()
						keys = slices.Grow(keys, int(n))[:n]
						_, err = p.Values().ReadValues(keys)
						if err != nil {
							if !errors.Is(err, io.EOF) {
								return err
							}
						}
					}
					{
						p, err := valuePages.ReadPage()
						if err != nil {
							return err
						}
						n := p.NumValues()
						values = slices.Grow(values, int(n))[:n]
						_, err = p.Values().ReadValues(keys)
						if err != nil {
							if !errors.Is(err, io.EOF) {
								return err
							}
						}
					}
					for i := range keys {
						x := keys[i]
						if x.RepetitionLevel() == 0 {
							if i != 0 {
								b.Append(true)
								bk.AppendStringValues(ks, nil)
								bv.AppendStringValues(vs, nil)
								ks = ks[:0]
								vs = vs[:0]
							}
						}
						ks = append(ks, x.String())
						vs = append(vs, values[i].String())
					}
					b.Append(true)
					bk.AppendStringValues(ks, nil)
					bv.AppendStringValues(vs, nil)
					ks = ks[:0]
					vs = vs[:0]
					return nil
				},
			}
		},
	}
}
