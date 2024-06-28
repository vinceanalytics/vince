package db

import (
	"math/bits"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
)

const (
	// Row ids used for boolean fields.
	falseRowID = uint64(0)
	trueRowID  = uint64(1)

	// BSI bits used to check existence & sign.
	bsiExistsBit = 0
	bsiSignBit   = 1
	bsiOffsetBit = 2

	falseRowOffset = 0 * shardwidth.ShardWidth // fragment row 0
	trueRowOffset  = 1 * shardwidth.ShardWidth // fragment row 1
)

func mutex(m *roaring.Bitmap, id uint64, value uint64) {
	m.Add(value*shardwidth.ShardWidth + (id % shardwidth.ShardWidth))
}

func bsi(m *roaring.Bitmap, id, value uint64) {
	fragmentColumn := id % shardwidth.ShardWidth
	m.DirectAdd(fragmentColumn)
	lz := bits.LeadingZeros64(value)
	row := uint64(2)
	for mask := uint64(0x1); mask <= 1<<(64-lz) && mask != 0; mask = mask << 1 {
		if value&mask > 0 {
			m.DirectAdd(row*shardwidth.ShardWidth + fragmentColumn)
		}
		row++
	}
}

func boolean(m *roaring.Bitmap, id uint64, value bool) {
	fragmentColumn := id % shardwidth.ShardWidth
	if value {
		m.DirectAdd(trueRowOffset + fragmentColumn)
	} else {
		m.DirectAdd(falseRowOffset + fragmentColumn)
	}
}

func rangeLT(tx *rbf.Tx, view string, shard uint64, predicate uint64, allowEquality bool) (*rows.Row, error) {
	if predicate == 1 && !allowEquality {
		predicate, allowEquality = 0, true
	}

	// Start with set of columns with values set.
	b, err := rowShard(tx, view, shard, bsiExistsBit)
	if err != nil {
		return nil, err
	}
	switch {
	case predicate == 0 && !allowEquality:
		// Match all negative integers.
		return rows.NewRow(), nil
	case predicate == 0 && allowEquality:
		// Match all integers that are either negative or 0.
		return rangeEQ(tx, view, shard, 0)
	default:
		return rangeLTUnsigned(tx, view, shard, b, 64, predicate, allowEquality)
	}
}

func rangeLTUnsigned(tx *rbf.Tx, view string, shard uint64, filter *rows.Row, bitDepth, predicate uint64, allowEquality bool) (*rows.Row, error) {
	switch {
	case uint64(bits.Len64(predicate)) > bitDepth:
		fallthrough
	case predicate == (1<<bitDepth)-1 && allowEquality:
		// This query matches all possible values.
		return filter, nil
	case predicate == (1<<bitDepth)-1 && !allowEquality:
		// This query matches everything that is not (1<<bitDepth)-1.
		matches := rows.NewRow()
		for i := uint64(0); i < bitDepth; i++ {
			row, err := rowShard(tx, view, shard, uint64(bsiOffsetBit+i))
			if err != nil {
				return nil, err
			}
			matches = matches.Union(filter.Difference(row))
		}
		return matches, nil
	case allowEquality:
		predicate++
	}

	// Compare intermediate bits.
	matched := rows.NewRow()
	remaining := filter
	for i := int(bitDepth - 1); i >= 0 && predicate > 0 && remaining.Any(); i-- {
		row, err := rowShard(tx, view, shard, uint64(bsiOffsetBit+i))
		if err != nil {
			return nil, err
		}
		zeroes := remaining.Difference(row)
		switch (predicate >> uint(i)) & 1 {
		case 1:
			// Match everything with a zero bit here.
			matched = matched.Union(zeroes)
			predicate &^= 1 << uint(i)
		case 0:
			// Discard everything with a one bit here.
			remaining = zeroes
		}
	}

	return matched, nil
}

func rangeGT(tx *rbf.Tx, view string, shard uint64, predicate uint64, allowEquality bool) (*rows.Row, error) {
	b, err := rowShard(tx, view, shard, bsiExistsBit)
	if err != nil {
		return nil, err
	}
	switch {
	case predicate == 0 && !allowEquality:
		// Match all positive numbers except zero.
		nonzero, err := rangeNEQ(tx, view, shard, 0)
		if err != nil {
			return nil, err
		}
		b = nonzero
		fallthrough
	case predicate == 0 && allowEquality:
		// Match all positive numbers.
		return b, nil
	default:
		// Match all positive numbers greater than the predicate.
		return rangeGTUnsigned(tx, view, shard, b, 64, uint64(predicate), allowEquality)
	}
}

func rangeGTUnsigned(tx *rbf.Tx, view string, shard uint64, filter *rows.Row, bitDepth, predicate uint64, allowEquality bool) (*rows.Row, error) {
prep:
	switch {
	case predicate == 0 && allowEquality:
		// This query matches all possible values.
		return filter, nil
	case predicate == 0 && !allowEquality:
		// This query matches everything that is not 0.
		matches := rows.NewRow()
		for i := uint64(0); i < bitDepth; i++ {
			row, err := rowShard(tx, view, shard, uint64(bsiOffsetBit+i))
			if err != nil {
				return nil, err
			}
			matches = matches.Union(filter.Intersect(row))
		}
		return matches, nil
	case !allowEquality && uint64(bits.Len64(predicate)) > bitDepth:
		// The predicate is bigger than the BSI width, so nothing can be bigger.
		return rows.NewRow(), nil
	case allowEquality:
		predicate--
		allowEquality = false
		goto prep
	}

	// Compare intermediate bits.
	matched := rows.NewRow()
	remaining := filter
	predicate |= (^uint64(0)) << bitDepth
	for i := int(bitDepth - 1); i >= 0 && predicate < ^uint64(0) && remaining.Any(); i-- {
		row, err := rowShard(tx, view, shard, uint64(bsiOffsetBit+i))
		if err != nil {
			return nil, err
		}
		ones := remaining.Intersect(row)
		switch (predicate >> uint(i)) & 1 {
		case 1:
			// Discard everything with a zero bit here.
			remaining = ones
		case 0:
			// Match everything with a one bit here.
			matched = matched.Union(ones)
			predicate |= 1 << uint(i)
		}
	}

	return matched, nil
}

func rangeBetween(tx *rbf.Tx, view string, shard uint64, predicateMin, predicateMax uint64) (*rows.Row, error) {
	b, err := rowShard(tx, view, shard, bsiExistsBit)
	if err != nil {
		return nil, err
	}

	switch {
	case predicateMin == predicateMax:
		return rangeEQ(tx, view, shard, predicateMin)
	default:
		return rangeBetweenUnsigned(tx, view, shard, b, predicateMin, predicateMax)
	}
}

func rangeBetweenUnsigned(tx *rbf.Tx, view string, shard uint64, filter *rows.Row, predicateMin, predicateMax uint64) (*rows.Row, error) {
	switch {
	case predicateMax > (1<<64)-1:
		// The upper bound cannot be violated.
		return rangeGTUnsigned(tx, view, shard, filter, 64, predicateMin, true)
	case predicateMin == 0:
		// The lower bound cannot be violated.
		return rangeLTUnsigned(tx, view, shard, filter, 64, predicateMax, true)
	}

	// Compare any upper bits which are equal.
	diffLen := bits.Len64(predicateMax ^ predicateMin)
	remaining := filter
	for i := int(64 - 1); i >= diffLen; i-- {
		row, err := rowShard(tx, view, shard, uint64(bsiOffsetBit+i))
		if err != nil {
			return nil, err
		}
		switch (predicateMin >> uint(i)) & 1 {
		case 1:
			remaining = remaining.Intersect(row)
		case 0:
			remaining = remaining.Difference(row)
		}
	}

	// Clear the bits we just compared.
	equalMask := (^uint64(0)) << diffLen
	predicateMin &^= equalMask
	predicateMax &^= equalMask

	var err error
	remaining, err = rangeGTUnsigned(tx, view, shard, remaining, uint64(diffLen), predicateMin, true)
	if err != nil {
		return nil, err
	}
	remaining, err = rangeLTUnsigned(tx, view, shard, remaining, uint64(diffLen), predicateMax, true)
	if err != nil {
		return nil, err
	}
	return remaining, nil
}

func rangeEQ(tx *rbf.Tx, view string, shard uint64, predicate uint64, filter ...*rows.Row) (*rows.Row, error) {
	// Start with set of columns with values set.
	b, err := rowShard(tx, view, shard, bsiExistsBit)
	if err != nil {
		return nil, err
	}
	if len(filter) > 0 {
		b = b.Intersect(filter[0])
	}
	bitDepth := bits.LeadingZeros64(predicate)
	// Filter any bits that don't match the current bit value.
	for i := int(bitDepth - 1); i >= 0; i-- {
		row, err := rowShard(tx, view, shard, uint64(bsiOffsetBit+i))
		if err != nil {
			return nil, err
		}
		bit := (predicate >> uint(i)) & 1
		if bit == 1 {
			b = b.Intersect(row)
		} else {
			b = b.Difference(row)
		}
	}
	return b, nil
}

func rangeNEQ(tx *rbf.Tx, view string, shard uint64, predicate uint64) (*rows.Row, error) {
	// Start with set of columns with values set.
	b, err := rowShard(tx, view, shard, bsiExistsBit)
	if err != nil {
		return nil, err
	}

	// Get the equal bitmap.
	eq, err := rangeEQ(tx, view, shard, predicate)
	if err != nil {
		return nil, err
	}

	// Not-null minus the equal bitmap.
	b = b.Difference(eq)

	return b, nil
}

func transpose(tx *rbf.Tx, o *roaring64.Bitmap, view string, shard uint64, columns *rows.Row) error {
	exists, err := rowShard(tx, view, shard, bsiExistsBit)
	if err != nil {
		return err
	}
	if columns != nil {
		exists = exists.Intersect(columns)
	}
	if !exists.Any() {
		// No relevant BSI values are present in this fragment.
		return nil
	}
	// Populate a map with the BSI data.
	data := make(map[uint64]uint64)
	mergeBits(exists, 0, data)
	for i := uint64(0); i < 64; i++ {
		bits, err := rowShard(tx, view, shard, bsiOffsetBit+uint64(i))
		if err != nil {
			return err
		}
		bits = bits.Intersect(exists)
		mergeBits(bits, 1<<i, data)
	}
	for _, val := range data {
		// Convert to two's complement and add base back to value.
		val = uint64((2*(int64(val)>>63) + 1) * int64(val&^(1<<63)))
		o.Add(val)
	}
	return nil
}

func shardRows(tx *rbf.Tx, view string, start uint64, filters ...roaring.BitmapFilter) ([]uint64, error) {
	r := roaring64.New()
	err := shardRowsBitmap(tx, view, start, r, filters...)
	if err != nil {
		return nil, err
	}
	return r.ToArray(), nil
}

func shardRowsBitmap(tx *rbf.Tx, view string, start uint64, b *roaring64.Bitmap, filters ...roaring.BitmapFilter) error {
	cb := func(row uint64) error {
		b.Add(row)
		return nil
	}
	startKey := rowToKey(start)
	filter := roaring.NewBitmapRowFilter(cb, filters...)
	return tx.ApplyFilter(view, startKey, filter)
}

func rowShard(tx *rbf.Tx, view string, shard uint64, rowID uint64) (*rows.Row, error) {
	data, err := tx.OffsetRange(view,
		shard*shardwidth.ShardWidth,
		rowID*shardwidth.ShardWidth,
		(rowID+1)*shardwidth.ShardWidth,
	)
	if err != nil {
		return nil, err
	}
	row := &rows.Row{
		Segments: []rows.RowSegment{
			rows.NewSegment(data, shard, true),
		},
	}
	row.InvalidateCount()
	return row, nil
}

// width of roaring containers is 2^16
const containerWidth = 1 << 16

func rowToKey(rowID uint64) (key uint64) {
	return rowID * (shardwidth.ShardWidth / containerWidth)
}

func mergeBits(bits *rows.Row, mask uint64, out map[uint64]uint64) {
	for _, v := range bits.Columns() {
		out[v] |= mask
	}
}
