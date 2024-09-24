package sets

import (
	"sort"

	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/cursor"
	"github.com/gernest/rbf/dsl/query"
	"github.com/gernest/roaring"
	"github.com/gernest/rows"
)

type RowIterator struct {
	cursor *rbf.Cursor
	shard  uint64
	rowIDs []uint64
	cur    int
	wrap   bool
}

var _ query.RowIterator = (*RowIterator)(nil)

func NewRowIterator(c *rbf.Cursor, shard uint64, wrap bool, filters ...roaring.BitmapFilter) (*RowIterator, error) {
	var r []uint64
	err := cursor.Rows(c, 0, func(row uint64) error {
		r = append(r, row)
		return nil
	}, filters...)
	if err != nil {
		return nil, err
	}
	return &RowIterator{
		cursor: c,
		shard:  shard,
		rowIDs: r,
		wrap:   wrap,
	}, nil
}

func (it *RowIterator) Seek(offset int64, whence int) (int64, error) {
	idx := sort.Search(len(it.rowIDs), func(i int) bool {
		return it.rowIDs[i] >= uint64(offset)
	})
	it.cur = idx
	return int64(idx), nil
}

func (it *RowIterator) Next() (r *rows.Row, rowID uint64, _ *int64, wrapped bool, err error) {
	if it.cur >= len(it.rowIDs) {
		if !it.wrap || len(it.rowIDs) == 0 {
			return nil, 0, nil, true, nil
		}
		it.Seek(0, 0)
		wrapped = true
	}
	id := it.rowIDs[it.cur]
	r, err = cursor.Row(it.cursor, it.shard, id)
	if err != nil {
		return r, rowID, nil, wrapped, err
	}

	rowID = id

	it.cur++
	return r, rowID, nil, wrapped, nil
}
