package staples

import (
	"reflect"
	"sync"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/vinceanalytics/vince/filters"
	"github.com/vinceanalytics/vince/index"
	"github.com/vinceanalytics/vince/logger"
)

type Index struct {
	Browser        *index.ColumnImpl
	BrowserVersion *index.ColumnImpl
	City           *index.ColumnImpl
	Country        *index.ColumnImpl
	Domain         *index.ColumnImpl
	EntryPage      *index.ColumnImpl
	ExitPage       *index.ColumnImpl
	Host           *index.ColumnImpl
	Event          *index.ColumnImpl
	Os             *index.ColumnImpl
	OsVersion      *index.ColumnImpl
	Path           *index.ColumnImpl
	Referrer       *index.ColumnImpl
	ReferrerSource *index.ColumnImpl
	Region         *index.ColumnImpl
	Screen         *index.ColumnImpl
	UtmCampaign    *index.ColumnImpl
	UtmContent     *index.ColumnImpl
	UtmMedium      *index.ColumnImpl
	UtmSource      *index.ColumnImpl
	UtmTerm        *index.ColumnImpl
	mapping        map[string]*index.ColumnImpl
	mu             sync.Mutex
}

func NewIndex() *Index {
	idx := &Index{
		Browser:        index.NewColIdx(),
		BrowserVersion: index.NewColIdx(),
		City:           index.NewColIdx(),
		Country:        index.NewColIdx(),
		Domain:         index.NewColIdx(),
		EntryPage:      index.NewColIdx(),
		ExitPage:       index.NewColIdx(),
		Host:           index.NewColIdx(),
		Event:          index.NewColIdx(),
		Os:             index.NewColIdx(),
		OsVersion:      index.NewColIdx(),
		Path:           index.NewColIdx(),
		Referrer:       index.NewColIdx(),
		ReferrerSource: index.NewColIdx(),
		Region:         index.NewColIdx(),
		Screen:         index.NewColIdx(),
		UtmCampaign:    index.NewColIdx(),
		UtmContent:     index.NewColIdx(),
		UtmMedium:      index.NewColIdx(),
		UtmSource:      index.NewColIdx(),
		UtmTerm:        index.NewColIdx(),
		mapping:        make(map[string]*index.ColumnImpl),
	}
	r := reflect.ValueOf(idx).Elem()
	typ := r.Type()
	for i := 0; i < r.NumField(); i++ {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}
		idx.mapping[Camel(f.Name)] = r.Field(i).Interface().(*index.ColumnImpl)
	}
	return idx
}

var _ index.Index = (*Index)(nil)

func (idx *Index) Index(r arrow.Record) (index.Full, error) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	o := make(map[string]*index.FullColumn)
	for i := 0; i < int(r.NumCols()); i++ {
		name := r.ColumnName(i)
		x, ok := idx.mapping[name]
		if !ok {
			continue
		}
		x.Index(r.Column(i).(*array.Dictionary))
		n, err := x.Build([]string{name})
		if err != nil {
			return nil, err
		}
		o[name] = n
	}
	return FullIndex(o), nil
}

type FullIndex map[string]*index.FullColumn

var _ index.Full = (*FullIndex)(nil)

func (idx FullIndex) Match(b *roaring.Bitmap, m []*filters.CompiledFilter) {
	for _, x := range m {
		v, ok := idx[x.Column]
		if !ok {
			logger.Fail("Missing index column", "column", x.Column)
		}
		b.And(v.Match(x))
	}
}

func (idx FullIndex) Size() (n uint64) {
	for _, v := range idx {
		n += v.Size()
	}
	return
}

func (idx FullIndex) Columns(f func(column index.Column) error) error {
	for _, v := range idx {
		err := f(v)
		if err != nil {
			return err
		}
	}
	return nil
}
