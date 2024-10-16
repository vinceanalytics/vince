package bsi

import (
	"bytes"
	"fmt"
	"math/bits"
	"runtime"
	"sync"
	"text/tabwriter"

	"github.com/vinceanalytics/vince/internal/roaring"
)

const (
	// Min64BitSigned - Minimum 64 bit value
	Min64BitSigned = -9223372036854775808
	// Max64BitSigned - Maximum 64 bit value
	Max64BitSigned = 9223372036854775807
)

type Bitmap = roaring.Bitmap

func NewBitmap() *Bitmap {
	return roaring.NewBitmap()
}

// BSI is at its simplest is an array of bitmaps that represent an encoded
// binary value.  The advantage of a BSI is that comparisons can be made
// across ranges of values whereas a bitmap can only represent the existence
// of a single value for a given column ID.  Another usage scenario involves
// storage of high cardinality values.
//
// It depends upon the bitmap libraries.  It is not thread safe, so
// upstream concurrency guards must be provided.
type BSI struct {
	Source Source
}

// NewBSI constructs a new BSI. Note that it is your responsibility to ensure that
// the min/max values are set correctly. Queries CompareValue, MinMax, etc. will not
// work correctly if the min/max values are not set correctly.
func NewBSI(maxValue int64, minValue int64) *BSI {
	src := new(Slice)
	bitsz := bits.Len64(uint64(minValue))
	if bits.Len64(uint64(maxValue)) > bitsz {
		bitsz = bits.Len64(uint64(maxValue))
	}
	for i := range bitsz + 1 {
		src.GetOrCreate(i)
	}
	return &BSI{Source: src}
}

// NewDefaultBSI constructs an auto-sized BSI
func NewDefaultBSI() *BSI {
	return NewBSI(int64(0), int64(0))
}

func (b *BSI) Reset() {
	if s, ok := b.Source.(Reset); ok {
		s.Reset()
	}
}

// GetExistenceBitmap returns a pointer to the underlying existence bitmap of the BSI
func (b *BSI) GetExistenceBitmap() *Bitmap {
	return b.ex()
}

// ValueExists tests whether the value exists.
func (b *BSI) ValueExists(columnID uint64) bool {
	return b.GetExistenceBitmap().Contains(columnID)
}

// GetCardinality returns a count of unique column IDs for which a value has been set.
func (b *BSI) GetCardinality() uint64 {
	return uint64(b.GetExistenceBitmap().GetCardinality())
}

// BitCount returns the number of bits needed to represent values.
func (b *BSI) BitCount() int {
	return b.Source.BitCount()
}

// SetValue sets a value for a given columnID.
func (b *BSI) SetValue(columnID uint64, value int64) {
	// If max/min values are set to zero then automatically determine bit array size
	minBits := bits.Len64(uint64(value))
	for i := range minBits {
		if uint64(value)&(1<<uint64(i)) > 0 {
			b.must(i).Set(columnID)
		} else {
			b.must(i).Remove(columnID)
		}
	}
	b.muex().Set(columnID)
}

// GetValue gets the value at the column ID. Second param will be false for non-existent values.
func (b *BSI) GetValue(columnID uint64) (value int64, exists bool) {
	exists = b.ex().Contains(columnID)
	if !exists {
		return
	}
	for i := 0; i < b.BitCount(); i++ {
		e := b.get(i)
		if e == nil {
			break
		}
		if b.get(i).Contains(columnID) {
			value |= 1 << i
		}
	}
	return
}

// Operation identifier
type Operation int

const (
	// LT less than
	LT Operation = 1 + iota
	// LE less than or equal
	LE
	// EQ equal
	EQ
	// GE greater than or equal
	GE
	// GT greater than
	GT
	// RANGE range
	RANGE
	// MIN find minimum
	MIN
	// MAX find maximum
	MAX
)

type task struct {
	bsi          *BSI
	op           Operation
	valueOrStart int64
	end          int64
}

// CompareValue compares value.
// Values should be in the range of the BSI (max, min).  If the value is outside the range, the result
// might erroneous.  The operation parameter indicates the type of comparison to be made.
// For all operations with the exception of RANGE, the value to be compared is specified by valueOrStart.
// For the RANGE parameter the comparison criteria is >= valueOrStart and <= end.
// The parallelism parameter indicates the number of CPU threads to be applied for processing.  A value
// of zero indicates that all available CPU resources will be potentially utilized.
func (b *BSI) CompareValue(parallelism int, op Operation, valueOrStart, end int64,
	foundSet *Bitmap) *Bitmap {
	if foundSet == nil {
		foundSet = b.ex()
	}
	if foundSet == nil {
		return nil
	}
	// protect result against race conditions
	var mu sync.Mutex
	result := make([]*roaring.Bitmap, 0, runtime.NumCPU())
	x := b.BitCount()
	startIsNegative := x == 64 && uint64(valueOrStart)&(1<<uint64(x-1)) > 0
	endIsNegative := x == 64 && uint64(end)&(1<<uint64(x-1)) > 0

	batch := foundSet.Batch()
	batch.Run(foundSet, func(idx int, next roaring.Next) {
		results := NewBitmap()

		defer func() {
			mu.Lock()
			result = append(result, results)
			mu.Unlock()
		}()

		for cID, ok := next(); ok; cID, ok = next() {
			eq1, eq2 := true, true
			lt1, lt2, gt1 := false, false, false
			j := b.BitCount() - 1
			isNegative := false
			if x == 64 {
				isNegative = b.get(j).Contains(cID)
				j--
			}
			compStartValue := valueOrStart
			compEndValue := end
			if isNegative != startIsNegative {
				compStartValue = ^valueOrStart + 1
			}
			if isNegative != endIsNegative {
				compEndValue = ^end + 1
			}
			for ; j >= 0; j-- {
				sliceContainsBit := b.get(j).Contains(cID)

				if uint64(compStartValue)&(1<<uint64(j)) > 0 {
					// BIT in value is SET
					if !sliceContainsBit {
						if eq1 {
							if (op == GT || op == GE || op == RANGE) && startIsNegative && !isNegative {
								gt1 = true
							}
							if op == LT || op == LE {
								if !startIsNegative || (startIsNegative == isNegative) {
									lt1 = true
								}
							}
							eq1 = false
							break
						}
					}
				} else {
					// BIT in value is CLEAR
					if sliceContainsBit {
						if eq1 {
							if (op == LT || op == LE) && isNegative && !startIsNegative {
								lt1 = true
							}
							if op == GT || op == GE || op == RANGE {
								if startIsNegative || (startIsNegative == isNegative) {
									gt1 = true
								}
							}
							eq1 = false
							if op != RANGE {
								break
							}
						}
					}
				}

				if op == RANGE && uint64(compEndValue)&(1<<uint64(j)) > 0 {
					// BIT in value is SET
					if !sliceContainsBit {
						if eq2 {
							if !endIsNegative || (endIsNegative == isNegative) {
								lt2 = true
							}
							eq2 = false
							if startIsNegative && !endIsNegative {
								break
							}
						}
					}
				} else if op == RANGE {
					// BIT in value is CLEAR
					if sliceContainsBit {
						if eq2 {
							if isNegative && !endIsNegative {
								lt2 = true
							}
							eq2 = false
							break
						}
					}
				}

			}

			switch op {
			case LT:
				if lt1 {
					results.Set(cID)
				}
			case LE:
				if lt1 || (eq1 && (!startIsNegative || (startIsNegative && isNegative))) {
					results.Set(cID)
				}
			case EQ:
				if eq1 {
					results.Set(cID)
				}
			case GE:
				if gt1 || (eq1 && (startIsNegative || (!startIsNegative && !isNegative))) {
					results.Set(cID)
				}
			case GT:
				if gt1 {
					results.Set(cID)
				}
			case RANGE:
				if (eq1 || gt1) && (eq2 || lt2) {
					results.Set(cID)
				}
			default:
				panic(fmt.Sprintf("Operation [%v] not supported here", op))
			}
		}
	})
	return roaring.FastOr(result...)
}

// Sum all values contained within the foundSet.   As a convenience, the cardinality of the foundSet
// is also returned (for calculating the average).

// Sum all values contained within the foundSet.   As a convenience, the cardinality of the foundSet
// is also returned (for calculating the average).
func (b *BSI) Sum(foundSet *Bitmap) (sum int64, count uint64) {
	count = uint64(foundSet.GetCardinality())
	for i := 0; i < b.BitCount(); i++ {
		e := b.get(i)
		if e == nil {
			break
		}
		sum += int64(roaring.AndCardinality(foundSet, e) << uint(i))
	}
	return
}

func (b *BSI) Extract(foundSet *Bitmap) map[uint64]int64 {
	ex := b.ex()
	if ex == nil {
		return map[uint64]int64{}
	}
	match := roaring.And(ex, foundSet)
	result := make(map[uint64]int64, match.GetCardinality())
	for i := 0; i < b.BitCount(); i++ {
		e := b.get(i)
		if e == nil {
			break
		}
		exists := roaring.And(e, match)
		exists.Each(func(value uint64) {
			result[value] |= 1 << i
		})
	}
	return result
}

func (b *BSI) Transpose(foundSet *Bitmap) *Bitmap {
	ex := b.ex()
	if ex == nil {
		return roaring.NewBitmap()
	}
	match := roaring.And(ex, foundSet)
	if match.IsEmpty() {
		return match
	}
	batch := match.Batch()
	src := b.Source.(KV)
	size := src.BitCount()
	bits := src[1:]
	var mu sync.Mutex
	result := make([]*Bitmap, 0, runtime.NumCPU())
	batch.Run(match, func(idx int, next roaring.Next) {
		r := NewBitmap()
		for col, ok := next(); ok; col, ok = next() {
			value := int64(0)
			for i := range size {
				if bits[i].Contains(col) {
					value |= 1 << i
				}
			}
			r.Set(uint64(value))
		}
		mu.Lock()
		result = append(result, r)
		mu.Unlock()
	})
	return roaring.FastOr(result...)
}

// We only perform Or on a and b. we don't want to modify a or b
// because there is a posibility a is read from buffer which may corrupt the backing slice..
func (a *BSI) Or(b *BSI) *BSI {
	o := NewDefaultBSI()
	for i := 0; i < a.BitCount(); i++ {
		na := a.get(i)
		nb := b.get(i)
		if na == nil && nb == nil {
			// reached maximum bits
			break
		}
		no := o.must(i)
		if na == nil {
			no.Or(nb)
			continue
		}
		if nb == nil {
			no.Or(na)
			continue
		}
		no.Or(roaring.FastOr(na, nb))
	}
	o.ex().Or(roaring.FastOr(a.ex(), b.ex()))
	return o
}

func (b *BSI) String() string {
	k := b.GetExistenceBitmap().ToArray()
	var o bytes.Buffer
	w := tabwriter.NewWriter(&o, 0, 0, 1, ' ', tabwriter.AlignRight)
	var tmp bytes.Buffer
	for i := range k {
		if i != 0 {
			tmp.WriteByte('\t')
		}
		fmt.Fprint(&tmp, k[i])
	}
	tmp.WriteByte('\t')
	tmp.WriteByte('\n')
	w.Write(tmp.Bytes())
	tmp.Reset()
	for i := range k {
		if i != 0 {
			tmp.WriteByte('\t')
		}
		v, _ := b.GetValue(k[i])
		fmt.Fprint(&tmp, v)
	}
	tmp.WriteByte('\t')
	tmp.WriteByte('\n')
	w.Write(tmp.Bytes())
	w.Flush()
	return o.String()
}

func (b *BSI) muex() *roaring.Bitmap {
	return b.Source.GetOrCreate(0)
}

func (b *BSI) ex() *roaring.Bitmap {
	return b.Source.Get(0)
}

func (b *BSI) must(i int) *roaring.Bitmap {
	return b.Source.GetOrCreate(i + 1)
}

func (b *BSI) get(i int) *roaring.Bitmap {
	return b.Source.Get(i + 1)
}
