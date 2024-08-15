package ro2

import (
	"math/bits"

	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

const (
	// BSI bits used to check existence & sign.
	bsiExistsBit = 0
	bsiSignBit   = 1
	bsiOffsetBit = 2
)

func (tx *Tx) Cmp(field, shard uint64, op roaring64.Operation,
	start, end int64) *roaring64.Bitmap {
	bitDepth := uint64(bits.Len64(uint64(max(start, end))))
	switch op {
	case roaring64.LT:
		return tx.rangeLT(field, shard, bitDepth, start, false)
	case roaring64.LE:
		return tx.rangeLT(field, shard, bitDepth, start, true)
	case roaring64.GT:
		return tx.rangeGT(field, shard, bitDepth, start, false)
	case roaring64.GE:
		return tx.rangeGT(field, shard, bitDepth, start, true)
	case roaring64.EQ:
		return tx.rangeEQ(field, shard, bitDepth, start)
	case roaring64.RANGE:
		return tx.rangeBetween(field, shard, bitDepth, start, end)
	default:
		return roaring64.New()
	}
}

func (tx *Tx) rangeLT(field, shard, bitDepth uint64, predicate int64, allowEquality bool) *roaring64.Bitmap {
	if predicate == 1 && !allowEquality {
		predicate, allowEquality = 0, true
	}

	// Start with set of columns with values set.
	b := tx.Row(shard, field, bsiExistsBit)
	sign := tx.Row(shard, field, bsiSignBit)
	upredicate := absInt64(predicate)

	switch {
	case predicate == 0 && !allowEquality:
		// Match all negative integers.
		b.And(sign)
		return b
	case predicate == 0 && allowEquality:
		// Match all integers that are either negative or 0.
		zeroes := tx.rangeEQ(field, shard, bitDepth, 0)
		b.And(sign)
		b.Or(zeroes)
		return b
	case predicate < 0:
		// Match all every negative number beyond the predicate.
		b.And(sign)
		return tx.rangeGTUnsigned(field, shard, b, bitDepth, upredicate, allowEquality)
	default:
		// Match positive numbers less than the predicate, and all negatives.
		pos := tx.rangeLTUnsigned(field, shard, roaring64.AndNot(b, sign), bitDepth, upredicate, allowEquality)
		b.And(sign)
		b.Or(pos)
		return b
	}
}

func (tx *Tx) rangeGT(field, shard uint64, bitDepth uint64, predicate int64, allowEquality bool) *roaring64.Bitmap {
	if predicate == -1 && !allowEquality {
		predicate, allowEquality = 0, true
	}

	b := tx.Row(shard, field, bsiExistsBit)
	// Create predicate without sign bit.
	upredicate := absInt64(predicate)

	sign := tx.Row(shard, field, bsiSignBit)
	switch {
	case predicate == 0 && !allowEquality:
		// Match all positive numbers except zero.
		nonzero := tx.rangeNEQ(field, shard, bitDepth, 0)
		b = nonzero
		fallthrough
	case predicate == 0 && allowEquality:
		// Match all positive numbers.
		b.AndNot(sign)
		return b
	case predicate >= 0:
		// Match all positive numbers greater than the predicate.
		b.AndNot(sign)
		return tx.rangeGTUnsigned(field, shard, b, bitDepth, upredicate, allowEquality)
	default:
		// Match all positives and greater negatives.
		neg := tx.rangeLTUnsigned(field, shard, roaring64.And(b, sign), bitDepth, upredicate, allowEquality)
		b.AndNot(sign)
		b.Or(neg)
		return b
	}
}

func (tx *Tx) rangeBetween(field, shard, bitDepth uint64, predicateMin, predicateMax int64) *roaring64.Bitmap {
	b := tx.Row(shard, field, bsiExistsBit)

	// Convert predicates to unsigned values.
	upredicateMin, upredicateMax := absInt64(predicateMin), absInt64(predicateMax)

	switch {
	case predicateMin == predicateMax:
		return tx.rangeEQ(field, shard, bitDepth, predicateMin)
	case predicateMin >= 0:
		// Handle positive-only values.
		r := tx.Row(shard, field, bsiSignBit)
		return tx.rangeBetweenUnsigned(field, shard, roaring64.AndNot(b, r), bitDepth, upredicateMin, upredicateMax)
	case predicateMax < 0:
		// Handle negative-only values. Swap unsigned min/max predicates.
		r := tx.Row(shard, field, bsiSignBit)
		b.And(r)
		return tx.rangeBetweenUnsigned(field, shard, b, bitDepth, upredicateMax, upredicateMin)
	default:
		// If predicate crosses positive/negative boundary then handle separately and union.
		r0 := tx.Row(shard, field, bsiSignBit)
		pos := tx.rangeLTUnsigned(field, shard, roaring64.AndNot(b, r0), bitDepth, upredicateMax, true)
		neg := tx.rangeLTUnsigned(field, shard, roaring64.And(b, r0), bitDepth, upredicateMin, true)
		return roaring64.Or(pos, neg)
	}
}

func (tx *Tx) rangeBetweenUnsigned(field, shard uint64, filter *roaring64.Bitmap, bitDepth, predicateMin, predicateMax uint64) *roaring64.Bitmap {
	switch {
	case predicateMax > (1<<64)-1:
		// The upper bound cannot be violated.
		return tx.rangeGTUnsigned(field, shard, filter, 64, predicateMin, true)
	case predicateMin == 0:
		// The lower bound cannot be violated.
		return tx.rangeLTUnsigned(field, shard, filter, 64, predicateMax, true)
	}

	// Compare any upper bits which are equal.
	diffLen := bits.Len64(predicateMax ^ predicateMin)
	remaining := filter
	for i := int(bitDepth - 1); i >= diffLen; i-- {
		row := tx.Row(shard, field, uint64(bsiOffsetBit+i))
		switch (predicateMin >> uint(i)) & 1 {
		case 1:
			remaining = roaring64.And(remaining, row)
		case 0:
			remaining = roaring64.AndNot(remaining, row)
		}
	}

	// Clear the bits we just compared.
	equalMask := (^uint64(0)) << diffLen
	predicateMin &^= equalMask
	predicateMax &^= equalMask

	remaining = tx.rangeGTUnsigned(field, shard, remaining, uint64(diffLen), predicateMin, true)
	remaining = tx.rangeLTUnsigned(field, shard, remaining, uint64(diffLen), predicateMax, true)
	return remaining
}

func (tx *Tx) rangeGTUnsigned(field, shard uint64, filter *roaring64.Bitmap, bitDepth, predicate uint64, allowEquality bool) *roaring64.Bitmap {
prep:
	switch {
	case predicate == 0 && allowEquality:
		// This query matches all possible values.
		return filter
	case predicate == 0 && !allowEquality:
		// This query matches everything that is not 0.
		matches := roaring64.New()
		for i := uint64(0); i < bitDepth; i++ {
			row := tx.Row(shard, field, uint64(bsiOffsetBit+i))
			matches = roaring64.Or(matches, roaring64.And(filter, row))
		}
		return matches
	case !allowEquality && uint64(bits.Len64(predicate)) > bitDepth:
		// The predicate is bigger than the BSI width, so nothing can be bigger.
		return roaring64.New()
	case allowEquality:
		predicate--
		allowEquality = false
		goto prep
	}

	// Compare intermediate bits.
	matched := roaring64.New()
	remaining := filter
	predicate |= (^uint64(0)) << bitDepth
	for i := int(bitDepth - 1); i >= 0 && predicate < ^uint64(0) && !remaining.IsEmpty(); i-- {
		row := tx.Row(shard, field, uint64(bsiOffsetBit+i))
		row.And(remaining)
		ones := roaring64.And(remaining, row)
		switch (predicate >> uint(i)) & 1 {
		case 1:
			// Discard everything with a zero bit here.
			remaining = ones
		case 0:
			matched = roaring64.Or(matched, ones)
			predicate |= 1 << uint(i)
		}
	}
	return matched
}

func (tx *Tx) rangeLTUnsigned(field, shard uint64, filter *roaring64.Bitmap, bitDepth, predicate uint64, allowEquality bool) *roaring64.Bitmap {
	switch {
	case uint64(bits.Len64(predicate)) > bitDepth:
		fallthrough
	case predicate == (1<<bitDepth)-1 && allowEquality:
		// This query matches all possible values.
		return filter
	case predicate == (1<<bitDepth)-1 && !allowEquality:
		// This query matches everything that is not (1<<bitDepth)-1.
		matches := roaring64.New()
		for i := uint64(0); i < bitDepth; i++ {
			row := tx.Row(shard, field, uint64(bsiOffsetBit+i))
			matches = roaring64.Or(matches, roaring64.AndNot(filter, row))
		}
		return matches
	case allowEquality:
		predicate++
	}

	// Compare intermediate bits.
	matched := roaring64.New()
	remaining := filter
	for i := int(bitDepth - 1); i >= 0 && predicate > 0 && !remaining.IsEmpty(); i-- {
		row := tx.Row(shard, field, uint64(bsiOffsetBit+i))
		zeroes := roaring64.AndNot(remaining, row)
		switch (predicate >> uint(i)) & 1 {
		case 1:
			matched = roaring64.Or(matched, zeroes)
			predicate &^= 1 << uint(i)
		case 0:
			// Discard everything with a one bit here.
			remaining = row
		}
	}

	return matched
}

func (tx *Tx) rangeNEQ(field, shard, bitDepth uint64, predicate int64) *roaring64.Bitmap {
	// Start with set of columns with values set.
	b := tx.Row(shard, field, bsiExistsBit)

	// Get the equal bitmap.
	eq := tx.rangeEQ(field, shard, bitDepth, predicate)

	b.AndNot(eq)
	return b
}

func (tx *Tx) rangeEQ(field, shard, bitDepth uint64, predicate int64) *roaring64.Bitmap {
	// Start with set of columns with values set.
	b := tx.Row(shard, field, bsiExistsBit)
	upredicate := absInt64(predicate)
	if uint64(bits.Len64(upredicate)) > bitDepth {
		// Predicate is out of range.
		return roaring64.New()
	}

	// Filter to only positive/negative numbers.
	sign := tx.Row(shard, field, bsiSignBit)
	if predicate < 0 {
		b.And(sign) // only negatives
	} else {
		b.AndNot(sign) // only positives
	}
	// Filter any bits that don't match the current bit value.
	for i := int(bitDepth - 1); i >= 0; i-- {
		row := tx.Row(shard, field, uint64(bsiOffsetBit+i))

		bit := (upredicate >> uint(i)) & 1

		if bit == 1 {
			b.And(row)
		} else {
			b.AndNot(row)
		}
	}
	return b
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
