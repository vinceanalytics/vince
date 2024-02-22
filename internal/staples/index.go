package staples

import (
	"sort"
	"unsafe"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/vinceanalytics/vince/internal/filters"
	"github.com/vinceanalytics/vince/internal/index"
	"github.com/vinceanalytics/vince/internal/logger"
)

type Index struct{}

func NewIndex() *Index {
	return new(Index)
}

var _ index.Index = (*Index)(nil)

func (idx *Index) Index(r arrow.Record) (index.Full, error) {
	cIdx := index.NewColIdx()
	defer cIdx.Release()
	o := make(map[string]*index.FullColumn)
	for i := 0; i < int(r.NumCols()); i++ {
		a := r.Column(i)
		if a.DataType().ID() != arrow.DICTIONARY || a.NullN() == a.Len() {
			// skip columns that only nulls. This happens for instance when geo ip is not
			// configured or cases of utm* properties
			continue
		}
		cIdx.Index(a.(*array.Dictionary))
		name := r.ColumnName(i)
		n, err := cIdx.Build(name)
		if err != nil {
			return nil, err
		}
		o[name] = n
		cIdx.Reset()
	}
	lo, hi := Timestamps(r)
	return NewFullIdx(o, uint64(lo), uint64(hi)), nil
}

func Timestamps(r arrow.Record) (lo, hi int64) {
	a := r.Column(0).(*array.Int64).Int64Values()
	if len(a) > 0 {
		// record is sorted by timestamp
		lo = a[0]
		hi = a[len(a)-1]
	}
	return
}

type FullIndex struct {
	m              map[string]*index.FullColumn
	keys           []string
	min, max, size uint64
}

var _ index.Full = (*FullIndex)(nil)

var baseIndexSize = uint64(unsafe.Sizeof(FullIndex{}))

func NewFullIdx(m map[string]*index.FullColumn, min, max uint64) *FullIndex {
	keys := make([]string, 0, len(m))
	n := baseIndexSize
	for k, v := range m {
		n += uint64(len(k) * 2)
		n += v.Size()
		keys = append(keys, k)
	}
	n += uint64(len(keys) * 2)
	sort.Strings(keys)
	return &FullIndex{keys: keys, m: m, min: min, max: max, size: n}
}

func (idx *FullIndex) CanIndex() bool {
	return true
}

func (idx *FullIndex) Match(b *roaring.Bitmap, m []*filters.CompiledFilter) {
	for _, x := range m {
		v, ok := idx.m[x.Column]
		if !ok {
			logger.Fail("Missing index column", "column", x.Column)
		}
		b.And(v.Match(x))
	}
}

func (idx *FullIndex) Size() (n uint64) {
	return idx.size
}

func (idx *FullIndex) Min() (n uint64) {
	return idx.min
}

func (idx *FullIndex) Max() (n uint64) {
	return idx.max
}

func (idx *FullIndex) Columns(f func(column index.Column) error) error {
	for _, v := range idx.keys {
		err := f(idx.m[v])
		if err != nil {
			return err
		}
	}
	return nil
}
