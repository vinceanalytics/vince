package index

import (
	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/filters"
)

type Part interface {
	Full
	ID() string
	Record() arrow.Record
	Release()
}

type Full interface {
	Columns(f func(column Column) error) error
	Match(b *roaring.Bitmap, m []*filters.CompiledFilter)
	Size() uint64
	Min() uint64
	Max() uint64
	CanIndex() bool
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
	FindGranules(resource string, start int64, end int64) []string
}

func Accept(min, max, start, end int64) bool {
	return min <= end && start <= max
}

func AcceptWith(min, max, start, end int64, fn func()) bool {
	if min <= end {
		if start <= max {
			fn()
		}
		return true
	}
	return false
}
