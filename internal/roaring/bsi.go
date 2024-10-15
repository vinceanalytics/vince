package roaring

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/bits"
	"reflect"
	"runtime"
	"slices"
	"sync"
	"text/tabwriter"
	"unsafe"
)

const (
	// Min64BitSigned - Minimum 64 bit value
	Min64BitSigned = -9223372036854775808
	// Max64BitSigned - Maximum 64 bit value
	Max64BitSigned = 9223372036854775807
)

// BSI is at its simplest is an array of bitmaps that represent an encoded
// binary value.  The advantage of a BSI is that comparisons can be made
// across ranges of values whereas a bitmap can only represent the existence
// of a single value for a given column ID.  Another usage scenario involves
// storage of high cardinality values.
//
// It depends upon the bitmap libraries.  It is not thread safe, so
// upstream concurrency guards must be provided.
type BSI struct {
	bA       []*Bitmap
	eBM      *Bitmap // Existence BitMap
	MaxValue int64
	MinValue int64
}

// NewBSI constructs a new BSI. Note that it is your responsibility to ensure that
// the min/max values are set correctly. Queries CompareValue, MinMax, etc. will not
// work correctly if the min/max values are not set correctly.
func NewBSI(maxValue int64, minValue int64) *BSI {
	bitsz := bits.Len64(uint64(minValue))
	if bits.Len64(uint64(maxValue)) > bitsz {
		bitsz = bits.Len64(uint64(maxValue))
	}
	ba := make([]*Bitmap, bitsz)
	for i := range ba {
		ba[i] = NewBitmap()
	}
	return &BSI{bA: ba, MaxValue: maxValue, MinValue: minValue, eBM: NewBitmap()}
}

// NewDefaultBSI constructs an auto-sized BSI
func NewDefaultBSI() *BSI {
	return NewBSI(int64(0), int64(0))
}

func (b *BSI) Each(f func(idx byte, bs *Bitmap) error) error {
	err := f(0, b.eBM)
	if err != nil {
		return err
	}
	for i := range b.bA {
		err := f(byte(i)+1, b.bA[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BSI) Reset() {
	clear(b.bA)
	b.bA = b.bA[:0]
	b.eBM.Reset()
	b.MaxValue = 0
	b.MinValue = 0
}

// GetExistenceBitmap returns a pointer to the underlying existence bitmap of the BSI
func (b *BSI) GetExistenceBitmap() *Bitmap {
	return b.eBM
}

// ValueExists tests whether the value exists.
func (b *BSI) ValueExists(columnID uint64) bool {
	return b.eBM.Contains(uint64(columnID))
}

// GetCardinality returns a count of unique column IDs for which a value has been set.
func (b *BSI) GetCardinality() uint64 {
	return uint64(b.eBM.GetCardinality())
}

// BitCount returns the number of bits needed to represent values.
func (b *BSI) BitCount() int {
	return len(b.bA)
}

// SetValue sets a value for a given columnID.
func (b *BSI) SetValue(columnID uint64, value int64) {
	// If max/min values are set to zero then automatically determine bit array size
	if b.MaxValue == 0 && b.MinValue == 0 {
		minBits := bits.Len64(uint64(value))
		for len(b.bA) < minBits {
			b.bA = append(b.bA, NewBitmap())
		}
	}

	for i := 0; i < b.BitCount(); i++ {
		if uint64(value)&(1<<uint64(i)) > 0 {
			b.bA[i].Set(columnID)
		} else {
			b.bA[i].Remove(columnID)
		}
	}
	b.eBM.Set(columnID)
}

// GetValue gets the value at the column ID. Second param will be false for non-existent values.
func (b *BSI) GetValue(columnID uint64) (value int64, exists bool) {
	exists = b.eBM.Contains(columnID)
	if !exists {
		return
	}
	for i := 0; i < b.BitCount(); i++ {
		if b.bA[i].Contains(columnID) {
			value |= 1 << i
		}
	}
	return
}

type action func(t *task, batch []uint64, resultsChan chan *Bitmap, wg *sync.WaitGroup)

func parallelExecutor(parallelism int, t *task, e action, foundSet *Bitmap) *Bitmap {

	var n int = parallelism
	if n == 0 {
		n = runtime.NumCPU()
	}

	resultsChan := make(chan *Bitmap, n)

	card := uint64(foundSet.GetCardinality())
	x := card / uint64(n)

	remainder := card - (x * uint64(n))
	var batch []uint64
	var wg sync.WaitGroup
	iter := foundSet.ManyIterator()
	for i := 0; i < n; i++ {
		if i == n-1 {
			batch = make([]uint64, x+remainder)
		} else {
			batch = make([]uint64, x)
		}
		iter.NextMany(batch)
		wg.Add(1)
		go e(t, batch, resultsChan, &wg)
	}

	wg.Wait()

	close(resultsChan)

	ba := make([]*Bitmap, 0)
	for bm := range resultsChan {
		ba = append(ba, bm)
	}
	return FastOr(ba...)
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

	comp := &task{bsi: b, op: op, valueOrStart: valueOrStart, end: end}
	if foundSet == nil {
		return parallelExecutor(parallelism, comp, compareValue, b.eBM)
	}
	return parallelExecutor(parallelism, comp, compareValue, foundSet)
}

func compareValue(e *task, batch []uint64, resultsChan chan *Bitmap, wg *sync.WaitGroup) {

	defer wg.Done()

	results := NewBitmap()

	x := e.bsi.BitCount()
	startIsNegative := x == 64 && uint64(e.valueOrStart)&(1<<uint64(x-1)) > 0
	endIsNegative := x == 64 && uint64(e.end)&(1<<uint64(x-1)) > 0

	for i := 0; i < len(batch); i++ {
		cID := batch[i]
		eq1, eq2 := true, true
		lt1, lt2, gt1 := false, false, false
		j := e.bsi.BitCount() - 1
		isNegative := false
		if x == 64 {
			isNegative = e.bsi.bA[j].Contains(cID)
			j--
		}
		compStartValue := e.valueOrStart
		compEndValue := e.end
		if isNegative != startIsNegative {
			compStartValue = ^e.valueOrStart + 1
		}
		if isNegative != endIsNegative {
			compEndValue = ^e.end + 1
		}
		for ; j >= 0; j-- {
			sliceContainsBit := e.bsi.bA[j].Contains(cID)

			if uint64(compStartValue)&(1<<uint64(j)) > 0 {
				// BIT in value is SET
				if !sliceContainsBit {
					if eq1 {
						if (e.op == GT || e.op == GE || e.op == RANGE) && startIsNegative && !isNegative {
							gt1 = true
						}
						if e.op == LT || e.op == LE {
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
						if (e.op == LT || e.op == LE) && isNegative && !startIsNegative {
							lt1 = true
						}
						if e.op == GT || e.op == GE || e.op == RANGE {
							if startIsNegative || (startIsNegative == isNegative) {
								gt1 = true
							}
						}
						eq1 = false
						if e.op != RANGE {
							break
						}
					}
				}
			}

			if e.op == RANGE && uint64(compEndValue)&(1<<uint64(j)) > 0 {
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
			} else if e.op == RANGE {
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

		switch e.op {
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
			panic(fmt.Sprintf("Operation [%v] not supported here", e.op))
		}
	}

	resultsChan <- results
}

// MinMax - Find minimum or maximum value.
func (b *BSI) MinMax(parallelism int, op Operation, foundSet *Bitmap) int64 {

	var n int = parallelism
	if n == 0 {
		n = runtime.NumCPU()
	}

	resultsChan := make(chan int64, n)

	card := uint64(foundSet.GetCardinality())
	x := card / uint64(n)

	remainder := card - (x * uint64(n))
	var batch []uint64
	var wg sync.WaitGroup
	iter := foundSet.ManyIterator()
	for i := 0; i < n; i++ {
		if i == n-1 {
			batch = make([]uint64, x+remainder)
		} else {
			batch = make([]uint64, x)
		}
		iter.NextMany(batch)
		wg.Add(1)
		go b.minOrMax(op, batch, resultsChan, &wg)
	}

	wg.Wait()

	close(resultsChan)
	var minMax int64
	if op == MAX {
		minMax = Min64BitSigned
	} else {
		minMax = Max64BitSigned
	}

	for val := range resultsChan {
		if (op == MAX && val > minMax) || (op == MIN && val <= minMax) {
			minMax = val
		}
	}
	return minMax
}

func (b *BSI) minOrMax(op Operation, batch []uint64, resultsChan chan int64, wg *sync.WaitGroup) {

	defer wg.Done()

	x := b.BitCount()
	var value int64 = Max64BitSigned
	if op == MAX {
		value = Min64BitSigned
	}

	for i := 0; i < len(batch); i++ {
		cID := batch[i]
		eq := true
		lt, gt := false, false
		j := b.BitCount() - 1
		var cVal int64
		valueIsNegative := uint64(value)&(1<<uint64(x-1)) > 0 && bits.Len64(uint64(value)) == 64
		isNegative := false
		if x == 64 {
			isNegative = b.bA[j].Contains(cID)
			if isNegative {
				cVal |= 1 << uint64(j)
			}
			j--
		}
		compValue := value
		if isNegative != valueIsNegative {
			compValue = ^value + 1
		}
		for ; j >= 0; j-- {
			sliceContainsBit := b.bA[j].Contains(cID)
			if sliceContainsBit {
				cVal |= 1 << uint64(j)
			}
			if uint64(compValue)&(1<<uint64(j)) > 0 {
				// BIT in value is SET
				if !sliceContainsBit {
					if eq {
						eq = false
						if op == MAX && valueIsNegative && !isNegative {
							gt = true
							break
						}
						if op == MIN && (!valueIsNegative || (valueIsNegative == isNegative)) {
							lt = true
						}
					}
				}
			} else {
				// BIT in value is CLEAR
				if sliceContainsBit {
					if eq {
						eq = false
						if op == MIN && isNegative && !valueIsNegative {
							lt = true
						}
						if op == MAX && (valueIsNegative || (valueIsNegative == isNegative)) {
							gt = true
						}
					}
				}
			}
		}
		if lt || gt {
			value = cVal
		}
	}

	resultsChan <- value
}

// Sum all values contained within the foundSet.   As a convenience, the cardinality of the foundSet
// is also returned (for calculating the average).

// Sum all values contained within the foundSet.   As a convenience, the cardinality of the foundSet
// is also returned (for calculating the average).
func (b *BSI) Sum(foundSet *Bitmap) (sum int64, count uint64) {
	count = uint64(foundSet.GetCardinality())
	for i := 0; i < b.BitCount(); i++ {
		sum += int64(foundSet.AndCardinality(b.bA[i]) << uint(i))
	}
	return
}

func (b *BSI) Extract(foundSet *Bitmap) map[uint64]int64 {
	match := And(b.eBM, foundSet)
	result := make(map[uint64]int64, match.GetCardinality())
	for i := 0; i < b.BitCount(); i++ {
		exists := And(b.bA[i], match)
		exists.Each(func(value uint64) {
			result[value] |= 1 << i
		})
	}
	return result
}

// We only perform Or on a and b. we don't want to modify a or b
// because there is a posibility a is read from buffer which may corrupt the backing slice..
func (a *BSI) Or(b *BSI) *BSI {
	bits := max(a.BitCount(), b.BitCount())
	ba := make([]*Bitmap, bits)

	ax := a.BitCount()
	bx := b.BitCount()
	for i := 0; i < bits; i++ {
		if ax > i && bx > i {
			ba[i] = FastOr(a.bA[i], b.bA[i])
			continue
		}
		if ax > i {
			ba[i] = a.bA[i]
			continue
		}
		if bx > i {
			ba[i] = b.bA[i]
			continue
		}
	}
	return &BSI{
		bA:  ba,
		eBM: FastOr(a.eBM, b.eBM),
	}
}

func NewBSIFromBuffer(data []byte) *BSI {
	off, data := chunkLast(data, 2)
	offIdx := binary.BigEndian.Uint16(off)
	offsetsData, data := chunkLast(data, int(offIdx))
	offsets := toUint32Slice(offsetsData)
	start := offsets[0]
	ba := make([]*Bitmap, 0, len(offsets)-1)
	for _, end := range offsets[1:] {
		ba = append(ba, FromBuffer(data[start:end]))
		start = end
	}
	return &BSI{
		bA:  ba,
		eBM: FromBuffer(data[:offsets[0]]),
	}
}

func chunkLast(data []byte, n int) (chunk, left []byte) {
	return data[len(data)-n:], data[:len(data)-n]
}

func (b *BSI) ToBuffer() []byte {
	_, data := b.ToBufferWith(nil, nil)
	return data
}

func (b *BSI) ToBufferWith(offsets []uint32, data []byte) ([]uint32, []byte) {
	if b.eBM.IsEmpty() {
		return []uint32{}, []byte{}
	}
	offsets = slices.Grow(offsets[:0], 1+b.BitCount())
	data = slices.Grow(data[:0], b.GetSizeInBytes())
	return b.Append(offsets, data)
}

func (b *BSI) Append(offsets []uint32, data []byte) ([]uint32, []byte) {
	if b.eBM.IsEmpty() {
		return []uint32{}, []byte{}
	}
	data = append(data, b.eBM.ToBuffer()...)
	offsets = append(offsets, uint32(len(data)))
	for i := range b.bA {
		data = append(data, b.bA[i].ToBuffer()...)
		offsets = append(offsets, uint32(len(data)))
	}
	offsetData := toBytes(offsets)
	data = append(data, offsetData...)
	data = binary.BigEndian.AppendUint16(data, uint16(len(offsetData)))
	return offsets, data
}

func toBytes(b []uint32) []byte {
	var bs []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	hdr.Len = len(b) * 4
	hdr.Cap = hdr.Len
	hdr.Data = uintptr(unsafe.Pointer(&b[0]))
	return bs
}

func toUint32Slice(b []byte) (result []uint32) {
	var u32s []uint32
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&u32s))
	hdr.Len = len(b) / 4
	hdr.Cap = hdr.Len
	hdr.Data = uintptr(unsafe.Pointer(&b[0]))
	return u32s
}

// GetSizeInBytes - the size in bytes of the data structure
func (b *BSI) GetSizeInBytes() int {
	if b.eBM.IsEmpty() {
		return 0
	}
	// at maximum we have 64 bits pul existence we get 65. This means maximum size
	// for offsets array is 260 (65 * 4) bytes. a uint16 is enough to represet this value.
	size := 2 + // size of offset as uint16
		((1 + // esistence bitmap offset
			b.BitCount()) * 4)
	size += b.eBM.GetSizeInBytes()
	for _, bm := range b.bA {
		size += bm.GetSizeInBytes()
	}
	return size
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
