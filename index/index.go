package index

import (
	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/vinceanalytics/vince/filters"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
)

type Full interface {
	Columns(f func(column Column) error) error
	Match(b *roaring.Bitmap, m []*filters.CompiledFilter)
	Size() uint64
	Min() uint64
	Max() uint64
}

type Column interface {
	NumRows() uint32
	Name() string
	Empty() bool
	Fst() []byte
	Bitmaps(f func(int, *roaring.Bitmap) error) error
}

type Index interface {
	Index(arrow.Record) (Full, error)
}

type Primary interface {
	Add(resource string, granule *v1.Granule)
}
