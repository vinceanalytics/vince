package index

import (
	"bytes"
	"slices"
	"sort"
	"sync"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/blevesearch/vellum"
	"github.com/vinceanalytics/vince/filters"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
)

type FullColumn struct {
	FST     []byte
	path    []string
	NumRows uint32
	bitmaps []*roaring.Bitmap
	fst     *vellum.FST
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
	case v1.Filter_equal:
		r, ok, _ := fst.Get(m.Value)
		if !ok {
			return new(roaring.Bitmap)
		}
		return f.bitmaps[r].Clone()
	case v1.Filter_not_equal:
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
	case v1.Filter_re_equal:
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
	case v1.Filter_re_not_equal:
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
	c.indexString(e, e.Dictionary().(*array.Binary))
	c.rows = uint32(e.Len())
}

func (c *ColumnImpl) indexString(d *array.Dictionary, a *array.Binary) {
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
