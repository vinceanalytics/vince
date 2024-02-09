package cold

import (
	"encoding/binary"
	"sync"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/vinceanalytics/vince/segment"
)

type Int64 struct {
	base
	array *array.Int64

	// This field can be quiet large. Use bitmaps that we can reuse ro avoid
	// excessive allocation.
	mapping map[int64]*roaring.Bitmap

	tern numericTerm
}

var _ segment.Field = (*Int64)(nil)

func NewInt64(name string, a *array.Int64) *Int64 {
	m := make(map[int64]*roaring.Bitmap)
	for i, v := range a.Int64Values() {
		r, ok := m[v]
		if !ok {
			r = get()
			m[v] = r
		}
		r.Add(uint32(i))
	}
	return &Int64{
		base: base{
			name: name,
			len:  len(m),
		},
		array:   a,
		mapping: m,
	}
}

func (i *Int64) EachTerm(vt segment.VisitTerm) {
	for k, v := range i.mapping {
		vt(i.tern.newTerm(k, v))
	}
}

func (i *Int64) Release() {
	for _, r := range i.mapping {
		release(r)
	}
	clear(i.mapping)
}

var roaringPool = &sync.Pool{New: func() any { return new(roaring.Bitmap) }}

func get() *roaring.Bitmap {
	return roaringPool.Get().(*roaring.Bitmap)
}

func release(b *roaring.Bitmap) {
	b.Clear()
	roaringPool.Put(b)
}

type numericTerm struct {
	term   [binary.MaxVarintLen64]byte
	n      int
	bitmap *roaring.Bitmap
}

func (n *numericTerm) newTerm(value int64, r *roaring.Bitmap) *numericTerm {
	n.n = binary.PutVarint(n.term[:], value)
	n.bitmap = r
	return n
}

var _ segment.FieldTerm = (*numericTerm)(nil)

func (t *numericTerm) Term() []byte { return t.term[:t.n] }

func (t *numericTerm) Frequency() int { return int(t.bitmap.GetCardinality()) }

func (t *numericTerm) EachLocation(vl segment.VisitLocation) {
	var loc Location
	t.bitmap.Iterate(func(x uint32) bool {
		loc.pos = int(x)
		vl(&loc)
		return true
	})
}
