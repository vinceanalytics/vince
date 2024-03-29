package index

import (
	"bytes"
	"slices"
	"sort"
	"sync"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/blevesearch/vellum"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/filters"
)

type FullColumn struct {
	fst     []byte
	name    string
	numRows uint32
	bitmaps []*roaring.Bitmap
	vellum  *vellum.FST
}

var _ Column = (*FullColumn)(nil)

func (f *FullColumn) Equal(c Column) bool {
	o, ok := c.(*FullColumn)
	if !ok {
		return ok
	}
	if f.name != c.Name() || f.numRows != c.NumRows() {
		return false
	}
	if !bytes.Equal(f.fst, o.fst) {
		return false
	}
	if len(f.bitmaps) != len(o.bitmaps) {
		return false
	}
	for i := range f.bitmaps {
		if !f.bitmaps[i].Equals(o.bitmaps[i]) {
			return false
		}
	}
	return true
}
func (f *FullColumn) Empty() bool {
	return len(f.fst) == 0
}

func (f *FullColumn) Fst() []byte {
	return f.fst
}

func (f *FullColumn) Name() string {
	return f.name
}

func (f *FullColumn) NumRows() uint32 {
	return f.numRows
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
		return f.equalMatch(fst, m)
	case v1.Filter_not_equal:
		b := f.equalMatch(fst, m)
		o := new(roaring.Bitmap)
		for i := uint32(0); i < f.numRows; i++ {
			if !b.Contains(i) {
				o.Add(i)
			}
		}
		return o
	case v1.Filter_re_equal:
		return f.reMatch(fst, m)
	case v1.Filter_re_not_equal:
		b := f.reMatch(fst, m)
		o := new(roaring.Bitmap)
		for i := uint32(0); i < f.numRows; i++ {
			if !b.Contains(i) {
				o.Add(i)
			}
		}
		return o
	default:
		return new(roaring.Bitmap)
	}
}

func (f *FullColumn) equalMatch(fst *vellum.FST, m *filters.CompiledFilter) *roaring.Bitmap {
	r, ok, _ := fst.Get(m.Value)
	if !ok {
		return new(roaring.Bitmap)
	}
	return f.bitmaps[r].Clone()
}

func (f *FullColumn) reMatch(fst *vellum.FST, m *filters.CompiledFilter) *roaring.Bitmap {
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
}

func (f *FullColumn) Load() *vellum.FST {
	if f.vellum != nil {
		return f.vellum
	}
	fst, err := vellum.Load(f.fst)
	if err != nil {
		// fst is crucial for operations. Everything relies on this. It is fatal
		// situation if we can't load this in memory
		panic("failed loading fst index in memory " + err.Error())
	}
	f.vellum = fst
	return fst
}

func (f *FullColumn) Size() (n uint64) {
	n = uint64(len(f.fst))
	for _, b := range f.bitmaps {
		n += b.GetSerializedSizeInBytes()
	}
	return
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
			c.keys = append(c.keys, name)
			c.values = append(c.values, b)
		}
		b.Add(uint32(i))
	}
	sort.Strings(c.keys)
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

func (c *ColumnImpl) Build(name string) (*FullColumn, error) {
	if len(c.mapping) == 0 {
		return &FullColumn{}, nil
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
		name:    name,
		fst:     bytes.Clone(c.buffer.Bytes()),
		bitmaps: slices.Clone(c.values),

		numRows: c.rows,
	}, nil
}
