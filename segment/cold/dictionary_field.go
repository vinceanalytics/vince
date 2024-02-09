package cold

import (
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/vinceanalytics/vince/segment"
)

type base struct {
	name string
	len  int
}

type DictionaryField struct {
	base
	array *array.Dictionary
	data  *array.Binary
	term  dictionaryTern
}

func NewDict(name string, a *array.Dictionary, b *array.Binary) *DictionaryField {
	return &DictionaryField{
		base: base{
			name: name,
			len:  a.Len(),
		},
		array: a,
		data:  b,
	}
}

var _ segment.Field = (*DictionaryField)(nil)

func (f *base) Name() string {
	return f.name
}

func (f *base) Length() int {
	return f.len
}

func (f *base) Value() []byte { return nil }

func (f *base) Index() bool          { return true }
func (f *base) Store() bool          { return true }
func (f *base) IndexDocValues() bool { return false }

func (f *DictionaryField) EachTerm(vt segment.VisitTerm) {
	for i := 0; i < f.array.Len(); i++ {
		vt(f.term.newTerm(f, i))
	}
}

func (f *DictionaryField) WalkIndex(idx int, cb func(int)) {
	resolved := f.array.GetValueIndex(idx)
	for i := idx; i < f.array.Len(); i++ {
		if f.array.GetValueIndex(i) == resolved {
			cb(i)
		}
	}
}

type dictionaryTern struct {
	field     *DictionaryField
	idx       int
	frequency int
}

func (t *dictionaryTern) newTerm(f *DictionaryField, idx int) *dictionaryTern {
	var freq int
	f.WalkIndex(idx, func(_ int) {
		freq++
	})
	t.field = f
	t.frequency = freq
	t.idx = idx
	return t
}

var _ segment.FieldTerm = (*dictionaryTern)(nil)

func (t *dictionaryTern) Term() []byte {
	return t.field.data.Value(t.field.array.GetValueIndex(t.idx))
}

func (t *dictionaryTern) Frequency() int { return t.frequency }

func (t *dictionaryTern) EachLocation(vl segment.VisitLocation) {
	resolved := t.field.array.GetValueIndex(t.idx)
	var loc Location
	for i := t.idx; i < t.field.array.Len(); i++ {
		if t.field.array.GetValueIndex(i) == resolved {
			loc.pos = i
			vl(&loc)
		}
	}
}

type Location struct {
	field string
	start int
	end   int
	pos   int
	size  int
}

var _ segment.Location = (*Location)(nil)

func (l *Location) Field() string { return l.field }
func (l *Location) Start() int    { return l.start }
func (l *Location) End() int      { return l.end }
func (l *Location) Pos() int      { return l.pos }
func (l *Location) Size() int     { return l.size }
