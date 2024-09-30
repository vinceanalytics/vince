/*
 * Copyright 2021 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sroar

import (
	"fmt"
	"math"
	"math/bits"
	"strings"
)

// container uses extra 4 []uint16 in the front as header.
// container[0] is used for storing the size of the container, expressed in Uint16.
// The container size cannot exceed the vicinity of 8KB. At 8KB, we switch from packed arrays to
// bitmaps. We can fit the entire uint16 worth of bitmaps in 8KB (2^16 / 8 = 8
// KB).

const (
	typeArray  uint16 = 0x00
	typeBitmap uint16 = 0x01

	// Container header.
	indexSize        int = 0
	indexType        int = 1
	indexCardinality int = 2
	// Index 2 and 3 is used for cardinality. We need 2 uint16s to store cardinality because
	// 2^16 will not fit in uint16.
	startIdx uint16 = 4

	minContainerSize = 64 // In Uint16.
	// Bitmap container can contain 2^16 integers. Each integer would use one bit to represent.
	// Given that our data is represented in []uint16s, that'd mean the size of container to store
	// it would be divided by 16.
	// 4 for header and 4096 for storing bitmap container. In Uint16.
	maxContainerSize = 4 + (1<<16)/16
)

func dataAt(data []uint16, i int) uint16 { return data[int(startIdx)+i] }

func incrCardinality(data []uint16) {
	cur := getCardinality(data)
	if cur+1 > math.MaxUint16 {
		data[indexCardinality+1] = 1
	} else {
		data[indexCardinality]++
	}
}

var invalidCardinality int = math.MaxUint16 + 10
var maxCardinality int = math.MaxUint16 + 1

func getCardinality(data []uint16) int {
	// This sum has to be done using two ints to avoid overflow.
	return int(data[indexCardinality]) + int(data[indexCardinality+1])
}

func setCardinality(data []uint16, c int) {
	if c > math.MaxUint16 {
		data[indexCardinality] = math.MaxUint16
		data[indexCardinality+1] = uint16(c - math.MaxUint16)
	} else {
		data[indexCardinality] = uint16(c)
		data[indexCardinality+1] = 0
	}
}

func zeroOutContainer(c []uint16) {
	switch c[indexType] {
	case typeArray:
		array(c).zeroOut()
	case typeBitmap:
		bitmap(c).zeroOut()
	}
}

func removeRangeContainer(c []uint16, lo, hi uint16) {
	switch c[indexType] {
	case typeArray:
		array(c).removeRange(lo, hi)
	case typeBitmap:
		bitmap(c).removeRange(lo, hi)
	}
}

func calculateAndSetCardinality(data []uint16) {
	if data[indexType] != typeBitmap {
		panic("Non-bitmap containers should always have cardinality set correctly")
	}
	b := bitmap(data)
	card := b.cardinality()
	setCardinality(b, card)
}

type array []uint16

// find returns the index of the first element >= x.
// The index is based on data portion of the container, ignoring startIdx.
// If the element > than all elements present, then N is returned where N = cardinality of the
// container.
func (c array) find(x uint16) int {
	N := getCardinality(c)
	for i := int(startIdx); i < int(startIdx)+N; i++ {
		if len(c) <= int(i) {
			panic(fmt.Sprintf("find: %d len(c) %d <= i %d\n", x, len(c), i))
		}
		if c[i] >= x {
			return int(i - int(startIdx))
		}
	}
	return N
}

func (c array) rank(x uint16) int {
	N := getCardinality(c)
	idx := c.find(x)
	if idx == N {
		return -1
	}
	return idx
}

func (c array) has(x uint16) bool {
	N := getCardinality(c)
	idx := c.find(x)
	if idx == N {
		return false
	}
	return c[int(startIdx)+idx] == x
}

func (c array) add(x uint16) bool {
	idx := c.find(x)
	N := getCardinality(c)
	offset := int(startIdx) + idx

	if int(idx) < N {
		if c[offset] == x {
			return false
		}
		// The entry at offset is the first entry, which is greater than x. Move it to the right.
		copy(c[offset+1:], c[offset:])
	}
	c[offset] = x
	incrCardinality(c)
	return true
}

func (c array) remove(x uint16) bool {
	idx := c.find(x)
	N := getCardinality(c)
	offset := int(startIdx) + idx

	if int(idx) < N {
		if c[offset] != x {
			return false
		}
		copy(c[offset:], c[offset+1:])
		setCardinality(c, N-1)
		return true
	}
	return false
}

func (c array) removeRange(lo, hi uint16) {
	if hi < lo {
		panic(fmt.Sprintf("args must satisfy lo <= hi, got lo: %d, hi: %d\n", lo, hi))
	}
	loIdx := c.find(lo)
	hiIdx := c.find(hi)

	st := int(startIdx)
	loVal := c[st+loIdx]
	N := getCardinality(c)

	// remove range doesn't intersect with any element in the array.
	if hi < loVal || loIdx == N {
		return
	}
	if hiIdx == N {
		if loIdx > 0 {
			c = c[:int(startIdx)+loIdx-1]
		} else {
			c = c[:int(startIdx)]
		}
		setCardinality(c, loIdx)
		return
	}
	if c[st+hiIdx] == hi {
		hiIdx++
	}
	copy(c[st+loIdx:], c[st+hiIdx:])
	setCardinality(c, N-hiIdx+loIdx)
}

func (c array) zeroOut() {
	setCardinality(c, 0)
}

// TODO: Figure out how memory allocation would work in these situations. Perhaps use allocator here?
func (c array) andArray(other array) []uint16 {
	min := min(getCardinality(c), getCardinality(other))

	setc := c.all()
	seto := other.all()

	out := make([]uint16, int(startIdx)+min+1)
	num := uint16(intersection2by2(setc, seto, out[startIdx:]))

	// Truncate out to how many values were found.
	out = out[:startIdx+num+1]
	out[indexType] = typeArray
	out[indexSize] = uint16(len(out))
	setCardinality(out, int(num))
	return out
}

// TODO: We can do this operation in-place on the src array.
func (c array) andNotArray(other array, buf []uint16) []uint16 {
	max := getCardinality(c)
	out := make([]uint16, int(startIdx)+max+1)

	andRes := array(c.andArray(other)).all()
	srcVals := array(c).all()
	num := uint16(difference(srcVals, andRes, out[startIdx:]))

	// Truncate out to how many values were found.
	out = out[:startIdx+num+1]
	out[indexType] = typeArray
	out[indexSize] = uint16(len(out))
	setCardinality(out, int(num))
	return out
}

func (c array) orArray(other array, buf []uint16, runMode int) []uint16 {
	// We ignore runInline for this call.

	max := getCardinality(c) + getCardinality(other)
	if max > 4096 {
		// Use bitmap container.
		out := bitmap(c.toBitmapContainer(buf))
		// For now, just keep it as a bitmap. No need to change if the
		// cardinality is smaller than 4096.
		out.orArray(other, nil, runMode|runInline)
		// Return out because out is pointing to buf. This would allow the
		// receiver to copy out.
		return out
	}

	// The output would be of typeArray.
	out := buf[:int(startIdx)+max]
	num := union2by2(c.all(), other.all(), out[startIdx:])
	out[indexType] = typeArray
	out[indexSize] = uint16(len(out))
	setCardinality(out, num)
	return out
}

var tmp = make([]uint16, 8192)

func (c array) andBitmap(other bitmap) []uint16 {
	out := make([]uint16, int(startIdx)+getCardinality(c)+2) // some extra space.
	out[indexType] = typeArray

	pos := startIdx
	for _, x := range c.all() {
		out[pos] = x
		pos += other.bitValue(x)
	}

	// Ensure we have at least one empty slot at the end.
	res := out[:pos+1]
	res[indexSize] = uint16(len(res))
	setCardinality(res, int(pos-startIdx))
	return res
}

// TODO: Write an optmized version of this function.
func (c array) andNotBitmap(other bitmap, buf []uint16) []uint16 {
	assert(len(buf) == maxContainerSize)
	res := array(buf)
	Memclr(res)
	res[indexSize] = 4
	for _, e := range c.all() {
		if !other.has(e) {
			res.add(e)
		}
	}
	return res
}

func (c array) isFull() bool {
	N := getCardinality(c)
	return int(startIdx)+N >= len(c)
}

func (c array) all() []uint16 {
	N := getCardinality(c)
	return c[startIdx : int(startIdx)+N]
}

func (c array) minimum() uint16 {
	N := getCardinality(c)
	if N == 0 {
		return 0
	}
	return c[startIdx]
}

func (c array) maximum() uint16 {
	N := getCardinality(c)
	if N == 0 {
		return 0
	}
	return c[int(startIdx)+N-1]
}

func (c array) toBitmapContainer(buf []uint16) []uint16 {
	if len(buf) == 0 {
		buf = make([]uint16, maxContainerSize)
	} else {
		assert(len(buf) == maxContainerSize)
		assert(len(buf) == copy(buf, empty))
	}

	b := bitmap(buf)
	b[indexSize] = maxContainerSize
	b[indexType] = typeBitmap
	setCardinality(b, getCardinality(c))

	data := b[startIdx:]
	for _, x := range c.all() {
		idx := x >> 4
		pos := x & 0xF
		data[idx] |= bitmapMask[pos]
	}
	return b
}

func (c array) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Size: %d\n", c[0]))
	for i, val := range c[startIdx:] {
		b.WriteString(fmt.Sprintf("%d: %d\n", i, val))
	}
	return b.String()
}

type bitmap []uint16

var bitmapMask []uint16

func init() {
	bitmapMask = make([]uint16, 16)
	for i := 0; i < 16; i++ {
		bitmapMask[i] = 1 << (15 - i)
	}
}

func (b bitmap) add(x uint16) bool {
	idx := x >> 4
	pos := x & 0xF

	if has := b[startIdx+idx] & bitmapMask[pos]; has > 0 {
		return false
	}

	b[startIdx+idx] |= bitmapMask[pos]
	incrCardinality(b)
	return true
}

func (b bitmap) remove(x uint16) bool {
	idx := x >> 4
	pos := x & 0xF

	c := getCardinality(b)
	if has := b[startIdx+idx] & bitmapMask[pos]; has > 0 {
		b[startIdx+idx] ^= bitmapMask[pos]
		setCardinality(b, c-1)
		return true
	}
	return false
}

func (b bitmap) removeRange(lo, hi uint16) {
	loIdx := lo >> 4
	loPos := lo & 0xF

	hiIdx := hi >> 4
	hiPos := hi & 0xF

	N := getCardinality(b)
	var removed int
	for i := loIdx + 1; i < hiIdx; i++ {
		removed += bits.OnesCount16(b[startIdx+i])
		b[startIdx+i] = 0
	}

	if loIdx == hiIdx {
		for p := loPos; p <= hiPos; p++ {
			if b[startIdx+loIdx]&bitmapMask[p] > 0 {
				removed++
			}
			b[startIdx+loIdx] &= ^bitmapMask[p]
		}
		setCardinality(b, N-removed)
		return
	}
	for p := loPos; p < 1<<4; p++ {
		if b[startIdx+loIdx]&bitmapMask[p] > 0 {
			removed++
		}
		b[startIdx+loIdx] &= ^bitmapMask[p]
	}
	for p := uint16(0); p <= hiPos; p++ {
		if b[startIdx+hiIdx]&bitmapMask[p] > 0 {
			removed++
		}
		b[startIdx+hiIdx] &= ^bitmapMask[p]
	}
	setCardinality(b, N-removed)
}

func (b bitmap) has(x uint16) bool {
	idx := x >> 4
	pos := x & 0xF
	has := b[startIdx+idx] & bitmapMask[pos]
	return has > 0
}

func (b bitmap) rank(x uint16) int {
	idx := x >> 4
	pos := x & 0xF
	if b[startIdx+idx]&bitmapMask[pos] == 0 {
		return -1
	}

	var rank int
	for i := 0; i < int(idx); i++ {
		rank += bits.OnesCount16(b[int(startIdx)+i])
	}
	for p := uint16(0); p <= pos; p++ {
		if b[startIdx+idx]&bitmapMask[p] > 0 {
			rank++
		}
	}
	return rank - 1
}

// TODO: This can perhaps be using SIMD instructions.
func (b bitmap) andBitmap(other bitmap) []uint16 {
	out := make([]uint16, maxContainerSize)
	out[indexSize] = maxContainerSize
	out[indexType] = typeBitmap
	var num int
	for i := int(startIdx); i < len(b); i++ {
		out[i] = b[i] & other[i]
		num += bits.OnesCount16(out[i])
	}
	setCardinality(out, num)
	return out
}

func (b bitmap) orBitmap(other bitmap, buf []uint16, runMode int) []uint16 {
	if runMode&runInline > 0 {
		buf = b
	} else {
		copy(buf, b) // Copy over first.
	}
	buf[indexSize] = maxContainerSize
	buf[indexType] = typeBitmap

	if num := getCardinality(b); num == maxCardinality {
		// do nothing. bitmap is already full.

	} else if runMode&runLazy > 0 || num == invalidCardinality {
		data := buf[startIdx:]
		for i, v := range other[startIdx:] {
			data[i] |= v
		}
		setCardinality(buf, invalidCardinality)

	} else {
		var num int
		data := buf[startIdx:]
		for i, v := range other[startIdx:] {
			data[i] |= v
			// We are going to iterate over the entire container. So, we can
			// just recount the cardinality, starting from num=0.
			num += bits.OnesCount16(data[i])
		}
		setCardinality(buf, num)
	}
	if runMode&runInline > 0 {
		return nil
	}
	return buf
}

func (b bitmap) andNotBitmap(other bitmap) []uint16 {
	var num int
	data := b[startIdx:]
	for i, v := range other[startIdx:] {
		data[i] = data[i] ^ (data[i] & v)
		num += bits.OnesCount16(data[i])
	}
	setCardinality(b, num)
	return b
}

func (b bitmap) andNotArray(other array) []uint16 {
	for _, e := range other.all() {
		b.remove(e)
	}
	return b
}

func (b bitmap) orArray(other array, buf []uint16, runMode int) []uint16 {
	if runMode&runInline > 0 {
		buf = b
	} else {
		copy(buf, b)
	}

	if num := getCardinality(b); num == maxCardinality {
		// do nothing. This bitmap is already full.

	} else if runMode&runLazy > 0 || num == invalidCardinality {
		// Avoid calculating the cardinality to speed up operations.
		for _, x := range other.all() {
			idx := x / 16
			pos := x % 16

			buf[startIdx+idx] |= bitmapMask[pos]
		}
		setCardinality(buf, invalidCardinality)

	} else {
		num := getCardinality(buf)
		for _, x := range other.all() {
			idx := x / 16
			pos := x % 16

			val := &buf[4+idx]
			before := bits.OnesCount16(*val)
			*val |= bitmapMask[pos]
			after := bits.OnesCount16(*val)
			num += after - before
		}
		setCardinality(buf, num)
	}

	if runMode&runInline > 0 {
		return nil
	}
	return buf
}

func (b bitmap) all() []uint16 {
	var res []uint16
	data := b[startIdx:]
	for idx := uint16(0); idx < uint16(len(data)); idx++ {
		x := data[idx]
		// TODO: This could potentially be optimized.
		for pos := uint16(0); pos < 16; pos++ {
			if x&bitmapMask[pos] > 0 {
				res = append(res, (idx<<4)|pos)
			}
		}
	}
	return res
}

//TODO: It can be optimized.
func (b bitmap) selectAt(idx int) uint16 {
	data := b[startIdx:]
	n := uint16(len(data))
	for i := uint16(0); i < n; i++ {
		x := data[i]
		c := bits.OnesCount16(x)
		if idx < c {
			for pos := uint16(0); pos < 16; pos++ {
				if idx == 0 && x&bitmapMask[pos] > 0 {
					return i*16 + pos
				}
				if x&bitmapMask[pos] > 0 {
					idx--
				}
			}

		}
		idx -= c
	}
	panic("should not reach here")
}

// bitValue returns a 0 or a 1 depending upon whether x is present in the bitmap, where 1 means
// present and 0 means absent.
func (b bitmap) bitValue(x uint16) uint16 {
	idx := x >> 4
	return (b[4+idx] >> (15 - (x & 0xF))) & 1
}

func (b bitmap) isFull() bool {
	return false
}

func (b bitmap) minimum() uint16 {
	N := getCardinality(b)
	if N == 0 {
		return 0
	}
	for i, x := range b[startIdx:] {
		lz := bits.LeadingZeros16(x)
		if lz == 16 {
			continue
		}
		return uint16(16*i + lz)
	}
	panic("We shouldn't reach here")
}

func (b bitmap) maximum() uint16 {
	N := getCardinality(b)
	if N == 0 {
		return 0
	}
	for i := len(b) - 1; i >= int(startIdx); i-- {
		x := b[i]
		tz := bits.TrailingZeros16(x)
		if tz == 16 {
			continue
		}
		return uint16(16*i + 15 - tz)
	}
	panic("We shouldn't reach here")
}

func (b bitmap) cardinality() int {
	var num int
	for _, x := range b[startIdx:] {
		num += bits.OnesCount16(x)
	}
	return num
}

var zeroContainer = make([]uint16, maxContainerSize)

func (b bitmap) zeroOut() {
	setCardinality(b, 0)
	copy(b[startIdx:], zeroContainer[startIdx:])
}

var (
	runInline = 0x01
	runLazy   = 0x02
)

func containerOr(ac, bc, buf []uint16, runMode int) []uint16 {
	at := ac[indexType]
	bt := bc[indexType]

	if at == typeArray && bt == typeArray {
		left := array(ac)
		right := array(bc)
		// We can't always inline this function. If the right container has
		// enough entries, trying to do a union with the left container inplace
		// could end up overwriting the left container entries. So, we use a
		// buffer to hold all output, and then copy it over to left.
		//
		// TODO: If right doesn't have a lot of entries, we could just iterate
		// over left and merge the entries from right inplace. Would be faster
		// than copying over all entries into buffer. Worth trying that approach.
		return left.orArray(right, buf, runMode)
	}
	if at == typeArray && bt == typeBitmap {
		left := array(ac)
		right := bitmap(bc)
		// Don't run inline for this call.
		return right.orArray(left, buf, runMode&^runInline)
	}

	// These two following cases can be fully inlined.
	if at == typeBitmap && bt == typeArray {
		left := bitmap(ac)
		right := array(bc)
		return left.orArray(right, buf, runMode)
	}
	if at == typeBitmap && bt == typeBitmap {
		left := bitmap(ac)
		right := bitmap(bc)
		return left.orBitmap(right, buf, runMode)
	}
	panic("containerOr: We should not reach here")
}

func containerAnd(ac, bc []uint16) []uint16 {
	at := ac[indexType]
	bt := bc[indexType]

	if at == typeArray && bt == typeArray {
		left := array(ac)
		right := array(bc)
		return left.andArray(right)
	}
	if at == typeArray && bt == typeBitmap {
		left := array(ac)
		right := bitmap(bc)
		return left.andBitmap(right)
	}
	if at == typeBitmap && bt == typeArray {
		left := bitmap(ac)
		right := array(bc)
		out := right.andBitmap(left)
		return out
	}
	if at == typeBitmap && bt == typeBitmap {
		left := bitmap(ac)
		right := bitmap(bc)
		return left.andBitmap(right)
	}
	panic("containerAnd: We should not reach here")
}

// TODO: Optimize this function.
func containerAndNot(ac, bc, buf []uint16) []uint16 {
	at := ac[indexType]
	bt := bc[indexType]

	if at == typeArray && bt == typeArray {
		left := array(ac)
		right := array(bc)
		return left.andNotArray(right, buf)
	}
	if at == typeArray && bt == typeBitmap {
		left := array(ac)
		right := bitmap(bc)
		return left.andNotBitmap(right, buf)
	}
	if at == typeBitmap && bt == typeArray {
		left := bitmap(ac)
		right := array(bc)
		out := left.andNotArray(right)
		return out
	}
	if at == typeBitmap && bt == typeBitmap {
		left := bitmap(ac)
		right := bitmap(bc)
		return left.andNotBitmap(right)
	}
	panic("containerAndNot: We should not reach here")
}
