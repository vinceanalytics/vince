package index

import (
	"bytes"
	"encoding/binary"
	"errors"
	"slices"
	"sort"
	"sync"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/blevesearch/vellum"
	"github.com/vinceanalytics/staples/staples/filters"
)

type FullColumn struct {
	FST     []byte
	path    []string
	NumRows uint32
	bitmaps []*roaring.Bitmap
	fst     *vellum.FST
	latest  *roaring.Bitmap
}

var _ Column = (*FullColumn)(nil)

func (f *FullColumn) Empty() bool {
	return len(f.FST) == 0
}

func (f *FullColumn) Fst() []byte {
	return f.FST
}

func (f *FullColumn) Path() []string {
	return f.path
}

func (f *FullColumn) Bitmaps(fn func(int, *roaring.Bitmap) error) error {
	for i := range f.bitmaps {
		if err := fn(i, f.bitmaps[i]); err != nil {
			return err
		}
	}
	return nil
}

func (f *FullColumn) Match(m *filters.CompiledFilter) *roaring.Bitmap {
	if f.Empty() {
		return new(roaring.Bitmap)
	}
	fst := f.Load()
	switch m.Op {
	case filters.Equal:
		r, ok, _ := fst.Get(m.Value)
		if !ok {
			return new(roaring.Bitmap)
		}
		return f.bitmaps[r].Clone()
	case filters.NotEqual:
		r, ok, _ := fst.Get(m.Value)
		if !ok {
			return new(roaring.Bitmap)
		}
		b := f.bitmaps[r]

		// any rows not in b bitmap
		o := new(roaring.Bitmap)
		for i := uint32(0); i < f.NumRows; i++ {
			if !b.Contains(i) {
				o.Add(i)
			}
		}
		return o
	case filters.ReMatch:
		var b *roaring.Bitmap
		it, err := fst.Search(m.Re, nil, nil)
		for err == nil {
			_, r := it.Current()
			if b == nil {
				b = f.bitmaps[r].Clone()
			} else {
				b.And(f.bitmaps[r])
			}
			err = it.Next()
		}
		if b == nil {
			b = new(roaring.Bitmap)
		}
		return b
	case filters.ReNotMatch:
		var b *roaring.Bitmap
		it, err := fst.Search(m.Re, nil, nil)
		for err == nil {
			_, r := it.Current()
			if b == nil {
				b = f.bitmaps[r].Clone()
			} else {
				b.And(f.bitmaps[r])
			}
			err = it.Next()
		}
		if b == nil {
			b = new(roaring.Bitmap)
		}
		o := new(roaring.Bitmap)
		for i := uint32(0); i < f.NumRows; i++ {
			if !b.Contains(i) {
				o.Add(i)
			}
		}
		return o
	case filters.Latest:
		return f.Latest()
	default:
		return new(roaring.Bitmap)
	}
}

func (f *FullColumn) Load() *vellum.FST {
	if f.fst != nil {
		return f.fst
	}
	fst, err := vellum.Load(f.FST)
	if err != nil {
		// fst is crucial for operations. Everything relies on this. It is fatal
		// situation if we can't load this in memory
		panic("failed loading fst index in memory " + err.Error())
	}
	f.fst = fst
	return fst
}

func (f *FullColumn) Latest() *roaring.Bitmap {
	if f.latest != nil {
		return f.latest
	}
	b := new(roaring.Bitmap)
	it, err := f.Load().Iterator(nil, nil)
	for err == nil {
		_, val := it.Current()
		b.Add(f.bitmaps[val].Maximum())
		err = it.Next()
	}
	if !errors.Is(err, vellum.ErrIteratorDone) {
		panic("invalid state of fst iteration " + err.Error())
	}
	f.latest = b
	return f.latest
}

func (f *FullColumn) Size() (n uint64) {
	n = uint64(len(f.FST))
	for _, b := range f.bitmaps {
		n += b.GetSerializedSizeInBytes()
	}
	return
}

type FullMapColumn map[string]*FullColumn

func (m FullMapColumn) Size() (n uint64) {
	for _, v := range m {
		n += v.Size()
	}
	return
}

func (f FullMapColumn) Match(m *filters.CompiledFilter) *roaring.Bitmap {
	c, ok := f[m.Column]
	if !ok {
		return new(roaring.Bitmap)
	}
	return c.Match(m)
}

type ColumnImpl struct {
	mapping map[string]*roaring.Bitmap
	rows    uint32
	keys    []string
	values  []*roaring.Bitmap
	buffer  bytes.Buffer
	build   *vellum.Builder
}

var columnPool = &sync.Pool{New: func() any { return newColIdx() }}

func (c *ColumnImpl) Index(e *array.Dictionary) {
	c.indexString(e, e.Dictionary().(*array.String))
	c.rows = uint32(e.Len())
}

func (c *ColumnImpl) indexString(d *array.Dictionary, a *array.String) {
	for i := 0; i < d.Len(); i++ {
		if d.IsNull(i) {
			continue
		}
		name := string(a.Value(d.GetValueIndex(i)))
		b, ok := c.mapping[name]
		if !ok {
			b = new(roaring.Bitmap)
			c.mapping[name] = b
		}
		b.Add(uint32(i))
	}
}

func NewColIdx() *ColumnImpl {
	return columnPool.Get().(*ColumnImpl)
}

func newColIdx() *ColumnImpl {
	col := &ColumnImpl{mapping: make(map[string]*roaring.Bitmap)}
	b, err := vellum.New(&col.buffer, nil)
	if err != nil {
		panic(err)
	}
	col.build = b
	return col
}

func (c *ColumnImpl) Release() {
	c.Reset()
	columnPool.Put(c)
}

func (c *ColumnImpl) Reset() {
	c.rows = 0
	clear(c.keys)
	clear(c.values)
	clear(c.mapping)
	c.buffer.Reset()
	c.keys = c.keys[:0]
	c.values = c.values[:0]
	c.build.Reset(&c.buffer)
}

func (c *ColumnImpl) Build(path []string) (*FullColumn, error) {
	if len(c.mapping) == 0 {
		return &FullColumn{}, nil
	}
	c.keys = slices.Grow(c.keys, len(c.mapping))
	c.values = slices.Grow(c.values, len(c.mapping))
	for k := range c.mapping {
		c.keys = append(c.keys, k)
	}
	sort.Strings(c.keys)
	for i := range c.keys {
		c.values = append(c.values, c.mapping[c.keys[i]])
	}
	for i := range c.keys {
		err := c.build.Insert([]byte(c.keys[i]), uint64(i))
		if err != nil {
			return nil, err
		}
	}
	err := c.build.Close()
	if err != nil {
		return nil, err
	}
	return &FullColumn{
		path:    path,
		FST:     bytes.Clone(c.buffer.Bytes()),
		bitmaps: slices.Clone(c.values),
		NumRows: c.rows,
	}, nil
}

type MapColumn map[string]*ColumnImpl

func NewMapIndex() MapColumn {
	return make(MapColumn)
}

func (m MapColumn) Build(path []string) (o FullMapColumn, err error) {
	o = make(FullMapColumn)
	for k, v := range m {
		f, err := v.Build(append(path, k))
		if err != nil {
			return nil, err
		}
		o[k] = f
	}
	return
}

func (m MapColumn) Reset() {
	for _, v := range m {
		v.Release()
	}
	clear(m)
}

func (m MapColumn) Index(ls *array.List) {
	rows := uint32(ls.Len())
	values := ls.ListValues()
	for i := 0; i < ls.Len(); i++ {
		if ls.IsNull(i) {
			continue
		}
		start, end := ls.ValueOffsets(i)
		chunk := array.NewSlice(values, start, end).(*array.Struct)
		key := chunk.Field(0).(*array.Dictionary)
		value := chunk.Field(1).(*array.Dictionary)
		m.addKeyValue(key, value, i, rows)
	}
}

func (m MapColumn) addKeyValue(key, value *array.Dictionary, row int, rows uint32) {
	kv := key.Dictionary().(*array.String)
	vv := value.Dictionary().(*array.String)
	for i := 0; i < key.Len(); i++ {
		column := kv.Value(key.GetValueIndex(i))
		c := m.get(column, rows)
		attr := vv.Value(value.GetValueIndex(i))
		x, ok := c.mapping[attr]
		if !ok {
			x = new(roaring.Bitmap)
			c.mapping[attr] = x
		}
		x.Add(uint32(row))
	}
}

func (m MapColumn) get(key string, rows uint32) *ColumnImpl {
	v, ok := m[key]
	if !ok {
		v = NewColIdx()
		v.rows = rows
		m[key] = v
	}
	return v
}

type Int64ColumnIndex struct {
	mapping map[uint64]*roaring.Bitmap
	rows    uint32
	keys    []uint64
	values  []*roaring.Bitmap
	buffer  bytes.Buffer
	build   *vellum.Builder
}

func (c *Int64ColumnIndex) Index(e *array.Uint64) {
	for i := 0; i < e.Len(); i++ {
		b, ok := c.mapping[e.Value(i)]
		if !ok {
			b = new(roaring.Bitmap)
			c.mapping[e.Value(i)] = b
		}
		b.Add(uint32(i))
	}
	c.rows = uint32(e.Len())
}

func NewInt64ColIdx() *Int64ColumnIndex {
	return int64ColumnPool.Get().(*Int64ColumnIndex)
}

var int64ColumnPool = &sync.Pool{New: func() any { return newInt64ColIdx() }}

func newInt64ColIdx() *Int64ColumnIndex {
	col := &Int64ColumnIndex{mapping: make(map[uint64]*roaring.Bitmap)}
	b, err := vellum.New(&col.buffer, nil)
	if err != nil {
		panic(err)
	}
	col.build = b
	return col
}

func (c *Int64ColumnIndex) Release() {
	c.Reset()
	int64ColumnPool.Put(c)
}

func (c *Int64ColumnIndex) Reset() {
	c.rows = 0
	clear(c.keys)
	clear(c.values)
	clear(c.mapping)
	c.buffer.Reset()
	c.keys = c.keys[:0]
	c.values = c.values[:0]
	c.build.Reset(&c.buffer)
}

func (c *Int64ColumnIndex) Build(path []string) (*FullColumn, error) {
	c.keys = slices.Grow(c.keys, len(c.mapping))
	c.values = slices.Grow(c.values, len(c.mapping))
	for k := range c.mapping {
		c.keys = append(c.keys, k)
	}
	slices.Sort(c.keys)
	for i := range c.keys {
		c.values = append(c.values, c.mapping[c.keys[i]])
	}
	var b [8]byte
	for i := range c.keys {
		binary.BigEndian.PutUint64(b[:], c.keys[i])
		err := c.build.Insert(b[:], uint64(i))
		if err != nil {
			return nil, err
		}
	}
	err := c.build.Close()
	if err != nil {
		return nil, err
	}
	return &FullColumn{
		path:    path,
		FST:     bytes.Clone(c.buffer.Bytes()),
		bitmaps: slices.Clone(c.values),
		NumRows: c.rows,
	}, nil
}
