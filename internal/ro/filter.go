package ro

import "fmt"

const (
	rowExponent = (Exponent - 16)     // for instance, 20-16 = 4
	rowWidth    = 1 << rowExponent    // containers per row, for instance 1<<4 = 16
	keyMask     = (rowWidth - 1)      // a mask for offset within the row
	rowMask     = ^FilterKey(keyMask) // a mask for the row bits, without converting them to a row ID
)

type FilterKey uint64

type Filter interface {
	ConsiderKey(key FilterKey, n int32) FilterResult
}

// FilterResult represents the results of a BitmapFilter considering a
// key, or data. The values are represented as exclusive upper bounds
// on a series of matches followed by a series of rejections. So for
// instance, if called on key 23, the result {YesKey: 23, NoKey: 24}
// indicates that key 23 is a "no" and 24 is unknown and will be the
// next to be Consider()ed.  This may seem confusing but it makes the
// math a lot easier to write. It can also report an error, which
// indicates that the entire operation should be stopped with that
// error.
type FilterResult struct {
	YesKey FilterKey // The lowest container key this filter is known NOT to match.
	NoKey  FilterKey // The highest container key after YesKey that this filter is known to not match.
	Err    error     // An error which should terminate processing.
}

// Row() computes the row number of a key.
func (f FilterKey) Row() uint64 {
	return uint64(f >> rowExponent)
}

// Add adds an offset to a key.
func (f FilterKey) Add(x uint64) FilterKey {
	return f + FilterKey(x)
}

// Sub determines the distance from o to f.
func (f FilterKey) Sub(o FilterKey) uint64 {
	return uint64(f - o)
}

// MatchReject just sets Yes and No appropriately.
func (f FilterKey) MatchReject(y, n FilterKey) FilterResult {
	return FilterResult{YesKey: y, NoKey: n}
}

func (f FilterKey) MatchOne() FilterResult {
	return FilterResult{YesKey: f + 1, NoKey: f + 1}
}

// NeedData() is only really meaningful for ConsiderKey, and indicates
// that a decision can't be made from the key alone.
func (f FilterKey) NeedData() FilterResult {
	return FilterResult{}
}

// Fail() reports a fatal error that should terminate processing.
func (f FilterKey) Fail(err error) FilterResult {
	return FilterResult{Err: err}
}

// Failf() is just like Errorf, etc
func (f FilterKey) Failf(msg string, args ...interface{}) FilterResult {
	return FilterResult{Err: fmt.Errorf(msg, args...)}
}

// MatchRow indicates that the current row matches the filter.
func (f FilterKey) MatchRow() FilterResult {
	return FilterResult{YesKey: (f & rowMask) + rowWidth}
}

// MatchOneRejectRow indicates that this item matched but no further
// items in this row can match.
func (f FilterKey) MatchOneRejectRow() FilterResult {
	return FilterResult{YesKey: f + 1, NoKey: (f & rowMask) + rowWidth}
}

// Reject rejects this item only.
func (f FilterKey) RejectOne() FilterResult {
	return FilterResult{NoKey: f + 1}
}

// Reject rejects N items.
func (f FilterKey) Reject(n uint64) FilterResult {
	return FilterResult{NoKey: f.Add(n)}
}

// RejectRow indicates that this entire row is rejected.
func (f FilterKey) RejectRow() FilterResult {
	return FilterResult{NoKey: (f & rowMask) + rowWidth}
}

// RejectUntil rejects everything up to the given key.
func (f FilterKey) RejectUntil(until FilterKey) FilterResult {
	return FilterResult{NoKey: until}
}

// RejectUntilRow rejects everything until the given row ID.
func (f FilterKey) RejectUntilRow(rowID uint64) FilterResult {
	return FilterResult{NoKey: FilterKey(rowID) << rowExponent}
}

// MatchRowUntilRow matches this row, then rejects everything else until
// the given row ID.
func (f FilterKey) MatchRowUntilRow(rowID uint64) FilterResult {
	// if rows are 16 wide, "yes" will be 16 minus our current position
	// within a row, and "no" will be the distance from the end of our
	// current row to the start of rowID, which is also the distance from
	// the beginning of our current row to the start of rowID-1.
	return FilterResult{
		YesKey: (f & rowMask) + rowWidth,
		NoKey:  FilterKey(rowID) << rowExponent,
	}
}

// RejectUntilOffset rejects this container, and any others until the given
// in-row offset.
func (f FilterKey) RejectUntilOffset(offset uint64) FilterResult {
	next := (f & rowMask).Add(offset)
	if next <= f {
		next += rowWidth
	}
	return FilterResult{NoKey: next}
}

// MatchUntilOffset matches the current container, then skips any other
// containers until the given offset.
func (f FilterKey) MatchOneUntilOffset(offset uint64) FilterResult {
	r := f.RejectUntilOffset(offset)
	r.YesKey = f + 1
	return r
}

// Done indicates that nothing can ever match.
func (f FilterKey) Done() FilterResult {
	return FilterResult{
		NoKey: ^FilterKey(0),
	}
}

// MatchRowAndDone matches this row and nothing after that.
func (f FilterKey) MatchRowAndDone() FilterResult {
	return FilterResult{
		YesKey: (f & rowMask) + rowWidth,
		NoKey:  ^FilterKey(0),
	}
}

// Match the current container, then skip any others until the same offset
// is reached again.
func (f FilterKey) MatchOneUntilSameOffset() FilterResult {
	return f.MatchOneUntilOffset(uint64(f) & keyMask)
}
