package bsi

import (
	"fmt"
	"slices"
	"sort"

	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/cursor"
	"github.com/gernest/rbf/dsl/query"
	"github.com/gernest/roaring"
	"github.com/gernest/rows"
)

type RowIterator struct {
	values []int64            // sorted slice of int values
	colIDs map[int64][]uint64 // [int value] -> [column IDs]
	cur    int                // current value index() (rowID)
	wrap   bool
}

var _ query.RowIterator = (*RowIterator)(nil)

func NewRowIterator(c *rbf.Cursor, shard uint64, wrap bool, filters ...roaring.BitmapFilter) (*RowIterator, error) {
	it := &RowIterator{
		colIDs: make(map[int64][]uint64),
		cur:    0,
		wrap:   wrap,
	}
	// accumulator [column ID] -> [int value]
	acc := make(map[uint64]int64)

	callback := func(rid uint64) error {
		// skip exist(0) and sign(1) rows
		if rid == bsiExistsBit || rid == bsiSignBit {
			return nil
		}
		val := int64(1 << (rid - bsiOffsetBit))
		r, err := cursor.Row(c, shard, rid)
		if err != nil {
			return err
		}
		for _, cid := range r.Columns() {
			acc[cid] |= val
		}
		return nil
	}
	if err := forEachRow(c, filters, callback); err != nil {
		return nil, err
	}

	// apply exist and sign bits
	r0, err := cursor.Row(c, shard, 0)
	if err != nil {
		return nil, err
	}
	allCols := r0.Columns()

	r1, err := cursor.Row(c, shard, 1)
	if err != nil {
		return nil, err
	}
	signCols := r1.Columns()
	signIdx, signLen := 0, len(signCols)
	// all distinct values
	values := make(map[int64]struct{})
	for _, cid := range allCols {
		// apply sign bit
		if signIdx < signLen && cid == signCols[signIdx] {
			if tmp, ok := acc[cid]; ok {
				acc[cid] = -tmp
			}

			signIdx++
		}

		val := acc[cid]
		it.colIDs[val] = append(it.colIDs[val], cid)

		if _, ok := values[val]; !ok {
			it.values = append(it.values, val)
			values[val] = struct{}{}
		}
	}
	slices.Sort(it.values)
	return it, nil
}

func (it *RowIterator) Seek(offset int64, whence int) (int64, error) {
	idx := sort.Search(len(it.values), func(i int) bool {
		return it.values[i] >= it.values[offset]
	})
	it.cur = idx
	return int64(idx), nil
}

func (it *RowIterator) Next() (r *rows.Row, rowID uint64, value *int64, wrapped bool, err error) {
	if it.cur >= len(it.values) {
		if !it.wrap || len(it.values) == 0 {
			return nil, 0, nil, true, nil
		}
		wrapped = true
		it.cur = 0
	}
	if it.cur >= 0 {
		rowID = uint64(it.cur)
		value = &it.values[rowID]
		r = rows.NewRow(it.colIDs[*value]...)
	}
	it.cur++
	return r, rowID, value, wrapped, nil
}

func forEachRow(c *rbf.Cursor, filters []roaring.BitmapFilter, fn func(rid uint64) error) error {
	filter := roaring.NewBitmapRowFilter(fn, filters...)
	err := c.ApplyFilter(0, filter)
	if err != nil {
		return fmt.Errorf("forEachRow %w", err)
	}
	return nil
}
