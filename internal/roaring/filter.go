package roaring

import (
	"errors"
	"fmt"
	"math/bits"
	"slices"
)

const (
	exponent                 = 20
	shardWidth               = 1 << exponent
	rowExponent              = (exponent - 16)     // for instance, 20-16 = 4
	rowWidth                 = 1 << rowExponent    // containers per row, for instance 1<<4 = 16
	keyMask                  = (rowWidth - 1)      // a mask for offset within the row
	rowMask                  = ^FilterKey(keyMask) // a mask for the row bits, without converting them to a row ID
	shardVsContainerExponent = 4
)

type FilterKey uint64

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

func (ra *Bitmap) ExtractMutex(match *Bitmap, f func(row uint64, columns *Bitmap)) {
	filter := make([][]uint16, 1<<shardVsContainerExponent)
	{
		iter := match.newCoIter()
		for iter.next() {
			k, c := iter.value()
			if getCardinality(c) == 0 {
				continue
			}
			filter[k%(1<<shardVsContainerExponent)] = c
		}
	}
	data := ra.newCoIter()
	prevRow := ^uint64(0)
	seenThisRow := false
	for data.next() {
		k, c := data.value()
		row := k >> shardVsContainerExponent
		if row == prevRow {
			if seenThisRow {
				continue
			}
		} else {
			seenThisRow = false
			prevRow = row
		}

		idx := k % (1 << shardVsContainerExponent)
		if len(filter[idx]) == 0 {
			continue
		}
		if containerAndAny(c, filter[idx]) {
			ex := containerAnd(c, filter[idx])
			f(row, toRows(ex))
			seenThisRow = true
		}
	}
}

func highbits(v uint64) uint64 { return v >> 16 }

func toRows(ac []uint16) *Bitmap {
	res := NewBitmap()
	offs := res.newContainer(uint16(len(ac)))
	copy(res.getContainer(offs), ac)
	res.setKey(0, offs)
	return res
}

func (ra *Bitmap) Mutex(id uint64, value uint64) {
	ra.Set(value*shardWidth + (id % shardWidth))
}

func (ra Bitmap) BSI(id uint64, svalue int64) {
	fragmentColumn := id % shardWidth
	ra.Set(fragmentColumn)
	negative := svalue < 0
	var value uint64
	if negative {
		ra.Set(shardWidth + fragmentColumn) // set sign bit
		value = uint64(svalue * -1)
	} else {
		value = uint64(svalue)
	}
	lz := bits.LeadingZeros64(value)
	row := uint64(2)
	for mask := uint64(0x1); mask <= 1<<(64-lz) && mask != 0; mask = mask << 1 {
		if value&mask > 0 {
			ra.Set(row*shardWidth + fragmentColumn)
		}
		row++
	}
}

func (ra *Bitmap) Row(shard, rowID uint64) *Bitmap {
	return ra.OffsetRange(
		shard*shardWidth,
		rowID*shardWidth,
		(rowID+1)*shardWidth,
	)
}

func (ra *Bitmap) OffsetRange(offset, start, end uint64) *Bitmap {
	keyEnd := end & mask
	off := offset & mask
	keyStart := start & mask
	res := NewBitmap()

	for i := ra.keys.search(start & mask); i < ra.keys.numKeys(); i++ {
		key := ra.keys.key(i)
		if key >= keyEnd {
			break
		}
		ac := ra.getContainer(ra.keys.val(i))
		offs := res.newContainer(uint16(len(ac)))
		copy(res.getContainer(offs), ac)
		res.setKey(off+(key-keyStart), offs)
	}
	return res
}

// ApplyFilterToIterator is a simplistic implementation that applies a bitmap
// filter to a ContainerIterator, returning an error if it encounters an error.
//
// This mostly exists for testing purposes; a Tx implementation where generating
// containers is expensive should almost certainly implement a better way to
// use filters which only generates data if it needs to.
func (ra *Bitmap) ApplyFilterToIterator(filter BitmapFilter) error {
	iter := ra.newCoIter()
	var until = uint64(0)
	for (until < ^uint64(0)) && iter.next() {
		key, data := iter.value()
		if key < until {
			continue
		}
		result := filter.ConsiderKey(FilterKey(key), int32(getCardinality(data)))
		if result.Err != nil {
			return result.Err
		}
		until = uint64(result.NoKey)
		if key < until {
			continue
		}
		result = filter.ConsiderData(FilterKey(key), data)
		if result.Err != nil {
			return result.Err
		}
		until = uint64(result.NoKey)
	}
	return nil
}

// BitmapFilter is, given a series of key/data pairs, considered to "match"
// some of those containers. Matching may be dependent on key values and
// cardinalities alone, or on the contents of the container.
//
// The ConsiderData function must not retain the container, or the data
// from the container; if it needs access to that information later, it needs
// to make a copy.
//
// Many filters are, by virtue of how they operate, able to predict their
// results on future keys. To accommodate this, and allow operations to
// avoid processing keys they don't need to process, the result of a filter
// operation can indicate not just whether a given key matches, but whether
// some upcoming keys will, or won't, match. If ConsiderKey yields a non-zero
// number of matches or non-matches for a given key, ConsiderData will not be
// called for that key.
//
// If multiple filters are combined, they are only called if their input is
// needed to determine a value.
type BitmapFilter interface {
	ConsiderKey(key FilterKey, n int32) FilterResult
	ConsiderData(key FilterKey, data []uint16) FilterResult
}

// BitmapColumnFilter is a BitmapFilter which checks for containers matching
// a given column within a row; thus, only the one container per row which
// matches the column needs to be evaluated, and it's evaluated as matching
// if it contains the relevant bit.
type BitmapColumnFilter struct {
	key, offset uint16
}

func NewBitmapColumnFilter(col uint64) BitmapFilter {
	return &BitmapColumnFilter{key: uint16((col >> 16) & keyMask), offset: uint16(col & 0xFFFF)}
}

var _ BitmapFilter = &BitmapColumnFilter{}

func (f *BitmapColumnFilter) ConsiderKey(key FilterKey, n int32) FilterResult {
	if uint16(key&keyMask) != f.key {
		return key.RejectUntilOffset(uint64(f.key))
	}
	return key.NeedData()
}

func (f *BitmapColumnFilter) ConsiderData(key FilterKey, data []uint16) FilterResult {
	if coContains(data, f.offset) {
		return key.MatchOneUntilSameOffset()
	}
	return key.RejectUntilOffset(uint64(f.key))
}

// BitmapRowsFilter is a BitmapFilter which checks for containers that are
// in any of a provided list of rows. The row list should be sorted.
type BitmapRowsFilter struct {
	rows []uint64
	i    int
}

func NewBitmapRowsFilter(rows []uint64) BitmapFilter {
	if len(rows) == 0 {
		return &BitmapRowsFilter{rows: rows, i: -1}
	}
	return &BitmapRowsFilter{rows: rows, i: 0}
}

func (f *BitmapRowsFilter) ConsiderKey(key FilterKey, n int32) FilterResult {
	if f.i == -1 {
		return key.Done()
	}
	if n == 0 {
		return key.RejectOne()
	}
	row := uint64(key) >> rowExponent
	for f.rows[f.i] < row {
		f.i++
		if f.i >= len(f.rows) {
			f.i = -1
			return key.Done()
		}
	}
	if f.rows[f.i] > row {
		return key.RejectUntilRow(f.rows[f.i])
	}
	// rows[f.i] must be equal, so we should match this row, until the
	// next row, if there is a next row.
	if f.i+1 < len(f.rows) {
		return key.MatchRowUntilRow(f.rows[f.i+1])
	}
	return key.MatchRowAndDone()
}

func (f *BitmapRowsFilter) ConsiderData(key FilterKey, data []uint16) FilterResult {
	return key.Fail(errors.New("bitmap rows filter should never consider data"))
}

// BitmapBSICountFilter gives counts of values in each value-holding row
// of a BSI field, constrained by a filter. The first row of the data is
// taken to be an existence bit, which is intersected into the filter to
// constrain it, and the second is used as a sign bit. The rows after that
// are treated as value rows, and their counts of bits, overlapping with
// positive and negative bits in the sign rows, are returned to a callback
// function.
//
// The total counts of positions evaluated are returned with a row count
// of ^uint64(0) prior to row counts.
type BitmapBSICountFilter struct {
	containers  [][]uint16
	positive    [][]uint16
	negative    [][]uint16
	nextOffsets []uint64
	count       int32
	psum, nsum  uint64
	buf         [maxContainerSize]uint16
}

// NewBitmapBSICountFilter creates a BitmapBSICountFilter, used for tasks
// like computing the sum of a BSI field matching a given filter.
//
// The input filter is assumed to represent one "row" of a shard's data,
// which is to say, a range of up to rowWidth consecutive containers starting
// at some multiple of rowWidth. We coerce that to the 0..rowWidth range
// because offset-within-row is what we care about.
func NewBitmapBSICountFilter(filter *Bitmap) *BitmapBSICountFilter {
	containers := make([][]uint16, rowWidth*3)
	b := &BitmapBSICountFilter{
		containers:  containers[:rowWidth],
		positive:    containers[rowWidth : rowWidth*2],
		negative:    containers[rowWidth*2 : rowWidth*3],
		nextOffsets: make([]uint64, rowWidth),
	}
	count := 0
	iter := filter.newCoIter()
	last := uint64(0)
	for iter.next() {
		k, v := iter.value()
		// Coerce container key into the 0-rowWidth range we'll be
		// using to compare against containers within each row.
		k = k & keyMask
		b.containers[k] = v
		last = k
		count++
	}
	// if there's only one container, we need to populate everything with
	// its position.
	if count == 1 {
		for i := range b.containers {
			b.nextOffsets[i] = last
		}
	} else {
		// Point each container at the offset of the next valid container.
		// With sparse bitmaps this will potentially make skipping faster.
		for i := range b.containers {
			if b.containers[i] != nil {
				for int(last) != i {
					b.nextOffsets[last] = uint64(i)
					last = (last + 1) % rowWidth
				}
			}
		}
	}
	return b
}

func (b *BitmapBSICountFilter) Total() (count int32, total int64) {
	return b.count, int64(b.psum) - int64(b.nsum)
}

func (b *BitmapBSICountFilter) ConsiderKey(key FilterKey, n int32) FilterResult {
	pos := key & keyMask
	if b.containers[pos] == nil || n == 0 {
		return key.RejectUntilOffset(b.nextOffsets[pos])
	}
	return key.NeedData()
}

func (b *BitmapBSICountFilter) ConsiderData(key FilterKey, data []uint16) FilterResult {
	pos := key & keyMask
	filter := b.containers[pos]
	if filter == nil {
		return key.RejectUntilOffset(b.nextOffsets[pos])
	}
	row := uint64(key >> rowExponent) // row count within the fragment
	// How do we translate the filter and existence bit into actionable things?
	// Assume the sign row is empty. We want positive values for anything in
	// the intersection of the filter and the positive bits. If the sign row
	// isn't empty, we want positive values for that intersection, less the
	// sign row, and negative for the intersection of the filter/positive and
	// the sign bits. So we can just stash the intermediate filter+existence
	// as positive, then split it up if we have sign bits, which we often don't.
	setup := false
	switch row {
	case 0: // existence bit
		b.positive[pos] = containerAnd(b.containers[pos], data)
		if &b.positive[pos][0] == &data[0] {
			b.positive[pos] = slices.Clone(b.positive[pos])
		}
		b.count += int32(getCardinality(b.positive[pos]))
		setup = true
	case 1: // sign bit
		// split into negative/positive components. doesn't affect total
		// count.
		b.negative[pos] = containerAnd(b.positive[pos], data)
		if &b.negative[pos][0] == &data[0] {
			b.negative[pos] = slices.Clone(b.negative[pos])
		}
		b.positive[pos] = containerAndNot(b.positive[pos], data, b.buf[:])
		setup = true
	}
	// if we were doing setup (first two rows), we're done
	if setup {
		return key.MatchOneUntilOffset(b.nextOffsets[pos])
	}
	// helpful reminder: a nil container is a valid empty container, and
	// intersectionCount knows this.
	pcount := containerAndCardinality(b.positive[pos], data)
	ncount := containerAndCardinality(b.negative[pos], data)
	b.psum += (uint64(pcount) << (row - 2))
	b.nsum += (uint64(ncount) << (row - 2))
	return key.MatchOneUntilOffset(b.nextOffsets[pos])
}

func coContains(ra []uint16, value uint16) bool {
	switch ra[indexType] {
	case typeArray:
		return array(ra).has(value)
	case typeBitmap:
		return bitmap(ra).has(value)
	default:
		return false
	}
}
