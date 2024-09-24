package bsi

import (
	"math/bits"

	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/cursor"
	"github.com/gernest/rows"
)

// Operation identifier
type Operation int

const (
	// LT less than
	LT Operation = 1 + iota
	// LE less than or equal
	LE
	// EQ equal
	EQ
	NEQ
	// GE greater than or equal
	GE
	// GT greater than
	GT
	// RANGE range
	RANGE
)

// Compare compares value.
// Values should be in the range of the BSI (max, min).  If the value is outside the range, the result
// might erroneous.  The operation parameter indicates the type of comparison to be made.
// For all operations with the exception of RANGE, the value to be compared is specified by valueOrStart.
// For the RANGE parameter the comparison criteria is >= valueOrStart and <= end.
//
// Returns column ID's satisfying the operation.
func Compare(
	c *rbf.Cursor,
	shard uint64, op Operation,
	valueOrStart int64, end int64,
	columns *rows.Row) (*rows.Row, error) {
	var r *rows.Row
	var err error
	bitDepth, err := depth(c)
	if err != nil {
		return nil, err
	}
	switch op {
	case LT:
		r, err = rangeLT(c, shard, bitDepth, valueOrStart, false)
	case LE:
		r, err = rangeLT(c, shard, bitDepth, valueOrStart, true)
	case GT:
		r, err = rangeGT(c, shard, bitDepth, valueOrStart, false)
	case GE:
		r, err = rangeGT(c, shard, bitDepth, valueOrStart, true)
	case EQ:
		r, err = rangeEQ(c, shard, bitDepth, valueOrStart)
	case NEQ:
		r, err = rangeNEQ(c, shard, bitDepth, valueOrStart)
	case RANGE:
		r, err = rangeBetween(c, shard, bitDepth, valueOrStart, end)
	default:
		return rows.NewRow(), nil
	}
	if err != nil {
		return nil, err
	}
	if columns != nil {
		r = r.Intersect(columns)
	}
	return r, nil
}

func rangeLT(c *rbf.Cursor, shard, bitDepth uint64, predicate int64, allowEquality bool) (*rows.Row, error) {
	if predicate == 1 && !allowEquality {
		predicate, allowEquality = 0, true
	}

	// Start with set of columns with values set.
	b, err := cursor.Row(c, shard, bsiExistsBit)
	if err != nil {
		return nil, err
	}
	sign, err := cursor.Row(c, shard, bsiSignBit)
	if err != nil {
		return nil, err
	}
	upredicate := absInt64(predicate)

	switch {
	case predicate == 0 && !allowEquality:
		// Match all negative integers.
		return b.Intersect(sign), nil
	case predicate == 0 && allowEquality:
		// Match all integers that are either negative or 0.
		zeroes, err := rangeEQ(c, shard, bitDepth, 0)
		if err != nil {
			return nil, err
		}
		return b.Intersect(sign).Union(zeroes), nil
	case predicate < 0:
		// Match all every negative number beyond the predicate.
		return rangeGTUnsigned(c, shard, b.Intersect(sign), bitDepth, upredicate, allowEquality)
	default:
		// Match positive numbers less than the predicate, and all negatives.
		pos, err := rangeLTUnsigned(c, shard, b.Difference(sign), bitDepth, upredicate, allowEquality)
		if err != nil {
			return nil, err
		}
		neg := b.Intersect(sign)
		return pos.Union(neg), nil
	}
}

func rangeLTUnsigned(c *rbf.Cursor, shard uint64, filter *rows.Row, bitDepth, predicate uint64, allowEquality bool) (*rows.Row, error) {
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
			row, err := cursor.Row(c, shard, uint64(bsiOffsetBit+i))
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
		row, err := cursor.Row(c, shard, uint64(bsiOffsetBit+i))
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

func rangeGT(c *rbf.Cursor, shard uint64, bitDepth uint64, predicate int64, allowEquality bool) (*rows.Row, error) {
	if predicate == -1 && !allowEquality {
		predicate, allowEquality = 0, true
	}

	b, err := cursor.Row(c, shard, bsiExistsBit)
	if err != nil {
		return nil, err
	}
	// Create predicate without sign bit.
	upredicate := absInt64(predicate)

	sign, err := cursor.Row(c, shard, bsiSignBit)
	if err != nil {
		return nil, err
	}
	switch {
	case predicate == 0 && !allowEquality:
		// Match all positive numbers except zero.
		nonzero, err := rangeNEQ(c, shard, bitDepth, 0)
		if err != nil {
			return nil, err
		}
		b = nonzero
		fallthrough
	case predicate == 0 && allowEquality:
		// Match all positive numbers.
		return b.Difference(sign), nil
	case predicate >= 0:
		// Match all positive numbers greater than the predicate.
		return rangeGTUnsigned(c, shard, b.Difference(sign), bitDepth, upredicate, allowEquality)
	default:
		// Match all positives and greater negatives.
		neg, err := rangeLTUnsigned(c, shard, b.Intersect(sign), bitDepth, upredicate, allowEquality)
		if err != nil {
			return nil, err
		}
		pos := b.Difference(sign)
		return pos.Union(neg), nil
	}
}

func rangeGTUnsigned(c *rbf.Cursor, shard uint64, filter *rows.Row, bitDepth, predicate uint64, allowEquality bool) (*rows.Row, error) {
prep:
	switch {
	case predicate == 0 && allowEquality:
		// This query matches all possible values.
		return filter, nil
	case predicate == 0 && !allowEquality:
		// This query matches everything that is not 0.
		matches := rows.NewRow()
		for i := uint64(0); i < bitDepth; i++ {
			row, err := cursor.Row(c, shard, uint64(bsiOffsetBit+i))
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
		row, err := cursor.Row(c, shard, uint64(bsiOffsetBit+i))
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

func rangeBetween(c *rbf.Cursor, shard, bitDepth uint64, predicateMin, predicateMax int64) (*rows.Row, error) {
	b, err := cursor.Row(c, shard, bsiExistsBit)
	if err != nil {
		return nil, err
	}

	// Convert predicates to unsigned values.
	upredicateMin, upredicateMax := absInt64(predicateMin), absInt64(predicateMax)

	switch {
	case predicateMin == predicateMax:
		return rangeEQ(c, shard, bitDepth, predicateMin)
	case predicateMin >= 0:
		// Handle positive-only values.
		r, err := cursor.Row(c, shard, bsiSignBit)
		if err != nil {
			return nil, err
		}
		return rangeBetweenUnsigned(c, shard, b.Difference(r), bitDepth, upredicateMin, upredicateMax)
	case predicateMax < 0:
		// Handle negative-only values. Swap unsigned min/max predicates.
		r, err := cursor.Row(c, shard, bsiSignBit)
		if err != nil {
			return nil, err
		}
		return rangeBetweenUnsigned(c, shard, b.Intersect(r), bitDepth, upredicateMax, upredicateMin)
	default:
		// If predicate crosses positive/negative boundary then handle separately and union.
		r0, err := cursor.Row(c, shard, bsiSignBit)
		if err != nil {
			return nil, err
		}
		pos, err := rangeLTUnsigned(c, shard, b.Difference(r0), bitDepth, upredicateMax, true)
		if err != nil {
			return nil, err
		}
		r1, err := cursor.Row(c, shard, bsiSignBit)
		if err != nil {
			return nil, err
		}
		neg, err := rangeLTUnsigned(c, shard, b.Intersect(r1), bitDepth, upredicateMin, true)
		if err != nil {
			return nil, err
		}
		return pos.Union(neg), nil
	}
}

func rangeBetweenUnsigned(c *rbf.Cursor, shard uint64, filter *rows.Row, bitDepth, predicateMin, predicateMax uint64) (*rows.Row, error) {
	switch {
	case predicateMax > (1<<64)-1:
		// The upper bound cannot be violated.
		return rangeGTUnsigned(c, shard, filter, 64, predicateMin, true)
	case predicateMin == 0:
		// The lower bound cannot be violated.
		return rangeLTUnsigned(c, shard, filter, 64, predicateMax, true)
	}

	// Compare any upper bits which are equal.
	diffLen := bits.Len64(predicateMax ^ predicateMin)
	remaining := filter
	for i := int(bitDepth - 1); i >= diffLen; i-- {
		row, err := cursor.Row(c, shard, uint64(bsiOffsetBit+i))
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
	remaining, err = rangeGTUnsigned(c, shard, remaining, uint64(diffLen), predicateMin, true)
	if err != nil {
		return nil, err
	}
	remaining, err = rangeLTUnsigned(c, shard, remaining, uint64(diffLen), predicateMax, true)
	if err != nil {
		return nil, err
	}
	return remaining, nil
}

func rangeEQ(c *rbf.Cursor, shard, bitDepth uint64, predicate int64) (*rows.Row, error) {
	// Start with set of columns with values set.
	b, err := cursor.Row(c, shard, bsiExistsBit)
	if err != nil {
		return nil, err
	}
	upredicate := absInt64(predicate)
	if uint64(bits.Len64(upredicate)) > bitDepth {
		// Predicate is out of range.
		return rows.NewRow(), nil
	}

	// Filter to only positive/negative numbers.
	sign, err := cursor.Row(c, shard, bsiSignBit)
	if err != nil {
		return nil, err
	}
	if predicate < 0 {
		b = b.Intersect(sign) // only negatives
	} else {
		b = b.Difference(sign) // only positives
	}
	// Filter any bits that don't match the current bit value.
	for i := int(bitDepth - 1); i >= 0; i-- {
		row, err := cursor.Row(c, shard, uint64(bsiOffsetBit+i))
		if err != nil {
			return nil, err
		}
		bit := (upredicate >> uint(i)) & 1

		if bit == 1 {
			b = b.Intersect(row)
		} else {
			b = b.Difference(row)
		}
	}

	return b, nil
}

func rangeNEQ(c *rbf.Cursor, shard, bitDepth uint64, predicate int64) (*rows.Row, error) {
	// Start with set of columns with values set.
	b, err := cursor.Row(c, shard, bsiExistsBit)
	if err != nil {
		return nil, err
	}

	// Get the equal bitmap.
	eq, err := rangeEQ(c, shard, bitDepth, predicate)
	if err != nil {
		return nil, err
	}

	// Not-null minus the equal bitmap.
	b = b.Difference(eq)

	return b, nil
}

func absInt64(v int64) uint64 {
	switch {
	case v > 0:
		return uint64(v)
	case v == -9223372036854775808:
		return 9223372036854775808
	default:
		return uint64(-v)
	}
}
