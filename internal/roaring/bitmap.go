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

package roaring

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

const mask = uint64(0xFFFFFFFFFFFF0000)

type Bitmap struct {
	data []uint16
	keys node
	_ptr []byte
	// memMoved keeps track of how many uint16 moves we had to do. The smaller
	// this number, the more efficient we have been.
	memMoved int
}

// FromBuffer returns a pointer to bitmap corresponding to the given buffer. This bitmap shouldn't
// be modified because it might corrupt the given buffer.
func FromBuffer(data []byte) *Bitmap {
	assert(len(data)%2 == 0)
	if len(data) < 8 {
		return NewBitmap()
	}
	du := toUint16Slice(data)
	x := toUint64Slice(du[:4])[indexNodeSize]
	return &Bitmap{
		data: du,
		// _ptr: data,
		keys: toUint64Slice(du[:x]),
	}
}

// FromBufferWithCopy creates a copy of the given buffer and returns a bitmap based on the copied
// buffer. This bitmap is safe for both read and write operations.
func FromBufferWithCopy(src []byte) *Bitmap {
	assert(len(src)%2 == 0)
	if len(src) < 8 {
		return NewBitmap()
	}
	src16 := toUint16Slice(src)
	dst16 := make([]uint16, len(src16))
	copy(dst16, src16)
	x := toUint64Slice(dst16[:4])[indexNodeSize]

	return &Bitmap{
		data: dst16,
		keys: toUint64Slice(dst16[:x]),
	}
}

func (ra *Bitmap) GetSizeInBytes() int {
	return len(ra.data) * 2
}

func (ra *Bitmap) ToBuffer() []byte {
	if ra.IsEmpty() {
		return nil
	}
	return toByteSlice(ra.data)
}

func (ra *Bitmap) ToBufferWithCopy() []byte {
	if ra.IsEmpty() {
		return nil
	}
	buf := make([]uint16, len(ra.data))
	copy(buf, ra.data)
	return toByteSlice(buf)
}

func NewBitmap() *Bitmap {
	return NewBitmapWith(2)
}

func NewBitmapWith(numKeys int) *Bitmap {
	if numKeys < 2 {
		panic("Must contain at least two keys.")
	}
	ra := &Bitmap{
		// Each key must also keep an offset. So, we need to double the number
		// of uint64s allocated. Plus, we need to make space for the first 2
		// uint64s to store the number of keys and node size.
		data: make([]uint16, 4*(2*numKeys+2)),
	}
	ra.keys = toUint64Slice(ra.data)
	ra.keys.setNodeSize(len(ra.data))

	// Always generate a container for key = 0x00. Otherwise, node gets confused
	// about whether a zero key is a new key or not.
	offset := ra.newContainer(minContainerSize)
	// First two are for num keys. index=2 -> 0 key. index=3 -> offset.
	ra.keys.setAt(indexNodeStart+1, offset)
	ra.keys.setNumKeys(1)

	return ra
}

func (ra *Bitmap) initSpaceForKeys(N int) {
	if N == 0 {
		return
	}
	curSize := uint64(len(ra.keys) * 4) // U64 -> U16
	bySize := uint64(N * 8)             // 2xU64 (key, value) -> 2x4xU16

	// The following code is borrowed from setKey.
	ra.scootRight(curSize, bySize)
	ra.keys = toUint64Slice(ra.data[:curSize+bySize])
	ra.keys.setNodeSize(int(curSize + bySize))
	assert(ra.keys.numKeys() == 1) // This initialization assumes that the number of keys are 1.

	// The containers have moved to the right bySize. So, update their offsets.
	// Currently, there's only one container.
	val := ra.keys.val(0)
	ra.keys.setAt(valOffset(0), val+uint64(bySize))
}

// setKey sets a key and container offset.
func (ra *Bitmap) setKey(k uint64, offset uint64) uint64 {
	if added := ra.keys.set(k, offset); !added {
		// No new key was added. So, we can just return.
		return offset
	}
	// A new key was added. Let's ensure that ra.keys is not full.
	if !ra.keys.isFull() {
		return offset
	}

	// ra.keys is full. We should expand its size.
	curSize := uint64(len(ra.keys) * 4) // Multiply by 4 for U64 -> U16.
	bySize := curSize
	if bySize > math.MaxUint16 {
		bySize = math.MaxUint16
	}

	ra.scootRight(curSize, bySize)
	ra.keys = toUint64Slice(ra.data[:curSize+bySize])
	ra.keys.setNodeSize(int(curSize + bySize))

	// All containers have moved to the right by bySize bytes.
	// Update their offsets.
	n := ra.keys
	for i := 0; i < n.maxKeys(); i++ {
		val := n.val(i)
		if val > 0 {
			n.setAt(valOffset(i), val+uint64(bySize))
		}
	}
	return offset + bySize
}

func (ra *Bitmap) fastExpand(bySize uint64) {
	prev := len(ra.keys) * 4 // Multiply by 4 to convert from u16 to u64.

	// This following statement also works. But, given how much fastExpand gets
	// called (a lot), probably better to control allocation.
	// ra.data = append(ra.data, empty[:bySize]...)

	toSize := len(ra.data) + int(bySize)
	if toSize <= cap(ra.data) {
		ra.data = ra.data[:toSize]
		return
	}
	growBy := cap(ra.data)
	if growBy < int(bySize) {
		growBy = int(bySize)
	}
	out := make([]uint16, cap(ra.data)+growBy)
	copy(out, ra.data)
	ra.data = out[:toSize]
	// Re-reference ra.keys correctly because underlying array has changed.
	ra.keys = toUint64Slice(ra.data[:prev])
}

// scootRight isn't aware of containers. It's going to create empty space of
// bySize at the given offset in ra.data. The offset doesn't need to line up
// with a container.
func (ra *Bitmap) scootRight(offset uint64, bySize uint64) {
	left := ra.data[offset:]

	ra.fastExpand(bySize) // Expand the buffer.
	right := ra.data[len(ra.data)-len(left):]
	n := copy(right, left) // Move data right.
	ra.memMoved += n

	Memclr(ra.data[offset : offset+uint64(bySize)]) // Zero out the space in the middle.
}

// scootLeft removes size number of uint16s starting from the given offset.
func (ra *Bitmap) scootLeft(offset uint64, size uint64) {
	n := uint64(len(ra.data))
	right := ra.data[offset+size:]
	ra.memMoved += copy(ra.data[offset:], right)
	ra.data = ra.data[:n-size]
}

func (ra *Bitmap) newContainer(sz uint16) uint64 {
	offset := uint64(len(ra.data))
	ra.fastExpand(uint64(sz))
	Memclr(ra.data[offset : offset+uint64(sz)])
	ra.data[offset] = sz
	return offset
}

// expandContainer would expand a container at the given offset. It would typically double the size
// of the container, until it reaches a threshold, where the size of the container would reach 2^16.
// Expressed in uint16s, that'd be (2^16)/(2^4) = 2^12 = 4096. So, if the container size >= 2048,
// then doubling that would put it above 4096. That's why in the code below, you see the checks for
// size 2048.
func (ra *Bitmap) expandContainer(offset uint64) {
	sz := ra.data[offset]
	if sz == 0 {
		panic("Container size should NOT be zero")
	}
	bySize := uint16(sz)
	if sz >= 2048 {
		// Size is in uint16. Half of max allowed size. If we're expanding the container by more
		// than 2048, we should just cap it to max size of 4096.
		assert(sz < maxContainerSize)
		bySize = maxContainerSize - sz
	}

	// Select the portion to the right of the container, beyond its right boundary.
	ra.scootRight(offset+uint64(sz), uint64(bySize))
	ra.keys.updateOffsets(offset, uint64(bySize), true)

	if sz < 2048 {
		ra.data[offset] = sz + bySize

	} else {
		// Convert to bitmap container.
		src := array(ra.getContainer(offset))
		buf := src.toBitmapContainer(nil)
		assert(copy(ra.data[offset:], buf) == maxContainerSize)
	}
}

// stepSize is used for container expansion. For a container of given size n,
// stepSize would return the target size. This function is used to reduce the
// number of times expansion needs to happen for each container.
func stepSize(n uint16) uint16 {
	// <=64 -> 128
	// <=128 -> 256
	// <=256 -> 512
	// <=512 -> 1024
	// <=1024 -> 2048
	// >1024 -> maxSize (convert to bitmap)
	for i := uint16(64); i <= 1024; i *= 2 {
		if n <= i {
			return i * 2
		}
	}
	return maxContainerSize
}

// copyAt would copy over a given container via src, into the container at
// offset. If src is a bitmap, it would copy it over directly. If src is an
// array container, then it would follow these paths:
//   - If src is smaller than dst, copy it over.
//   - If not, look for target size for dst using the stepSize function.
//   - If target size is maxSize, then convert src to a bitmap container, and
//     copy to dst.
//   - If target size is not max size, then expand dst container and copy src.
func (ra *Bitmap) copyAt(offset uint64, src []uint16) {
	dstSize := ra.data[offset]
	if dstSize == 0 {
		panic("Container size should NOT be zero")
	}

	// The src is a bitmapContainer. Just copy it over.
	if src[indexType] == typeBitmap {
		assert(src[indexSize] == maxContainerSize)
		bySize := uint16(maxContainerSize) - dstSize
		// Select the portion to the right of the container, beyond its right boundary.
		ra.scootRight(offset+uint64(dstSize), uint64(bySize))
		ra.keys.updateOffsets(offset, uint64(bySize), true)
		assert(copy(ra.data[offset:], src) == len(src))
		return
	}

	// src is an array container. Check if dstSize >= src. If so, just copy.
	// But, do keep dstSize intact, otherwise we'd lose portion of our container.
	if dstSize >= src[indexSize] {
		assert(copy(ra.data[offset:], src) == len(src))
		ra.data[offset] = dstSize
		return
	}

	// dstSize < src. Determine the target size of the container.
	targetSz := stepSize(dstSize)
	for targetSz < src[indexSize] {
		targetSz = stepSize(targetSz)
	}

	if targetSz == maxContainerSize {
		// Looks like the targetSize is now maxSize. So, convert src to bitmap container.
		s := array(src)

		bySize := uint16(maxContainerSize) - dstSize
		// Select the portion to the right of the container, beyond its right boundary.
		ra.scootRight(offset+uint64(dstSize), uint64(bySize))
		ra.keys.updateOffsets(offset, uint64(bySize), true)

		// Update the space of the container, so getContainer would work correctly.
		ra.data[offset] = maxContainerSize

		// Convert the src array to bitmap and write it directly over to the container.
		out := ra.getContainer(offset)
		Memclr(out)
		s.toBitmapContainer(out)
		return
	}

	// targetSize is not maxSize. Let's expand to targetSize and copy array.
	bySize := targetSz - dstSize
	ra.scootRight(offset+uint64(dstSize), uint64(bySize))
	ra.keys.updateOffsets(offset, uint64(bySize), true)
	assert(copy(ra.data[offset:], src) == len(src))
	ra.data[offset] = targetSz
}

func (ra Bitmap) getContainer(offset uint64) []uint16 {
	data := ra.data[offset:]
	if len(data) == 0 {
		panic(fmt.Sprintf("No container found at offset: %d\n", offset))
	}
	sz := data[0]
	return data[:sz]
}

func (ra *Bitmap) Clone() *Bitmap {
	abuf := ra.ToBuffer()
	bbuf := make([]byte, len(abuf))
	copy(bbuf, abuf)
	return FromBuffer(bbuf)
}

func (ra *Bitmap) IsEmpty() bool {
	if ra == nil {
		return true
	}
	N := ra.keys.numKeys()
	for i := 0; i < N; i++ {
		offset := ra.keys.val(i)
		cont := ra.getContainer(offset)
		if c := getCardinality(cont); c > 0 {
			return false
		}
	}
	return true
}

func (ra *Bitmap) Set(x uint64) bool {
	key := x & mask
	offset, has := ra.keys.getValue(key)
	if !has {
		// We need to add a container.
		o := ra.newContainer(minContainerSize)
		// offset might have been updated by setKey.
		offset = ra.setKey(key, o)
	}
	c := ra.getContainer(offset)
	switch c[indexType] {
	case typeArray:
		p := array(c)
		if added := p.add(uint16(x)); !added {
			return false
		}
		if p.isFull() {
			ra.expandContainer(offset)
		}
		return true
	case typeBitmap:
		b := bitmap(c)
		return b.add(uint16(x))
	}
	panic("we shouldn't reach here")
}

func FromSortedList(vals []uint64) *Bitmap {
	var arr []uint16
	var hi, lastHi, off uint64

	ra := NewBitmap()

	if len(vals) == 0 {
		return ra
	}

	// Set the keys beforehand so that we don't need to move a lot of memory because of adding keys.
	var numKeys int
	for _, x := range vals {
		hi = x & mask
		if hi != 0 && hi != lastHi {
			numKeys++
		}
		lastHi = hi
	}
	ra.initSpaceForKeys(numKeys)

	finalize := func(l []uint16, key uint64) {
		if len(l) == 0 {
			return
		}
		if len(l) <= 2048 {
			// 4 uint16s for the header, and extra 4 uint16s so that adding more elements using
			// Set operation doesn't fail.
			sz := uint16(8 + len(l))
			off = ra.newContainer(sz)
			c := ra.getContainer(off)
			c[indexSize] = sz
			c[indexType] = typeArray
			setCardinality(c, len(l))
			for i := 0; i < len(l); i++ {
				c[int(startIdx)+i] = l[i]
			}

		} else {
			off = ra.newContainer(maxContainerSize)
			c := ra.getContainer(off)
			c[indexSize] = maxContainerSize
			c[indexType] = typeBitmap
			for _, v := range l {
				bitmap(c).add(v)
			}
		}
		ra.setKey(key, off)
		return
	}

	lastHi = 0
	for _, x := range vals {
		hi = x & mask
		// Finalize the last container before proceeding ahead
		if hi != 0 && hi != lastHi {
			finalize(arr, lastHi)
			arr = arr[:0]
		}
		arr = append(arr, uint16(x))
		lastHi = hi
	}
	finalize(arr, lastHi)
	return ra
}

// TODO: Potentially this can be optimized.
func (ra *Bitmap) SetMany(vals []uint64) {
	for _, k := range vals {
		ra.Set(k)
	}
}

// Select returns the element at the xth index. (0-indexed)
func (ra *Bitmap) Select(x uint64) (uint64, error) {
	if x >= uint64(ra.GetCardinality()) {
		return 0, errors.Errorf("index %d is not less than the cardinality: %d",
			x, ra.GetCardinality())
	}
	n := ra.keys.numKeys()
	for i := 0; i < n; i++ {
		off := ra.keys.val(i)
		con := ra.getContainer(off)
		c := uint64(getCardinality(con))
		assert(c != uint64(invalidCardinality))
		if x < c {
			key := ra.keys.key(i)
			switch con[indexType] {
			case typeArray:
				return key | uint64(array(con).all()[x]), nil
			case typeBitmap:
				return key | uint64(bitmap(con).selectAt(int(x))), nil
			}
		}
		x -= c
	}
	panic("should not reach here")
}

func (ra *Bitmap) Contains(x uint64) bool {
	if ra == nil {
		return false
	}
	key := x & mask
	offset, has := ra.keys.getValue(key)
	if !has {
		return false
	}
	y := uint16(x)

	c := ra.getContainer(offset)
	switch c[indexType] {
	case typeArray:
		p := array(c)
		return p.has(y)
	case typeBitmap:
		b := bitmap(c)
		return b.has(y)
	}
	return false
}

func (ra *Bitmap) Remove(x uint64) bool {
	if ra == nil {
		return false
	}
	key := x & mask
	offset, has := ra.keys.getValue(key)
	if !has {
		return false
	}
	c := ra.getContainer(offset)
	switch c[indexType] {
	case typeArray:
		p := array(c)
		return p.remove(uint16(x))
	case typeBitmap:
		b := bitmap(c)
		return b.remove(uint16(x))
	}
	return true
}

// Remove range removes [lo, hi) from the bitmap.
func (ra *Bitmap) RemoveRange(lo, hi uint64) {
	if lo > hi {
		panic("lo should not be more than hi")
	}
	if lo == hi {
		return
	}

	k1 := lo & mask
	k2 := hi & mask

	defer ra.Cleanup()

	//  Complete range lie in a single container
	if k1 == k2 {
		if off, has := ra.keys.getValue(k1); has {
			c := ra.getContainer(off)
			removeRangeContainer(c, uint16(lo), uint16(hi)-1)
		}
		return
	}

	// Remove all the containers in range [k1+1, k2-1].
	n := ra.keys.numKeys()
	st := ra.keys.search(k1)
	key := ra.keys.key(st)
	if key == k1 {
		st++
	}

	for i := st; i < n; i++ {
		key := ra.keys.key(i)
		if key >= k2 {
			break
		}
		if off, has := ra.keys.getValue(key); has {
			zeroOutContainer(ra.getContainer(off))
		}
	}

	// Remove elements >= lo in k1's container
	if off, has := ra.keys.getValue(k1); has {
		c := ra.getContainer(off)
		if uint16(lo) == 0 {
			zeroOutContainer(c)
		} else {
			removeRangeContainer(c, uint16(lo), math.MaxUint16)
		}
	}

	if uint16(hi) == 0 {
		return
	}

	// Remove all elements < hi in k2's container
	if off, has := ra.keys.getValue(k2); has {
		c := ra.getContainer(off)
		removeRangeContainer(c, 0, uint16(hi)-1)
	}
}

func (ra *Bitmap) Reset() {
	// reset ra.data to size enough for one container and corresponding key.
	// 2 u64 is needed for header and another 2 u16 for the key 0.
	ra.data = ra.data[:16+minContainerSize]
	ra.keys = toUint64Slice(ra.data)

	offset := ra.newContainer(minContainerSize)
	ra.keys.setAt(indexNodeStart+1, offset)
	ra.keys.setNumKeys(1)
}

func (ra *Bitmap) GetCardinality() int {
	if ra == nil {
		return 0
	}
	N := ra.keys.numKeys()
	var sz int
	for i := 0; i < N; i++ {
		offset := ra.keys.val(i)
		c := ra.getContainer(offset)
		sz += getCardinality(c)
	}
	return sz
}

func (ra *Bitmap) ToArray() []uint64 {
	if ra == nil {
		return nil
	}
	res := make([]uint64, 0, ra.GetCardinality())
	N := ra.keys.numKeys()
	for i := 0; i < N; i++ {
		key := ra.keys.key(i)
		off := ra.keys.val(i)
		c := ra.getContainer(off)

		switch c[indexType] {
		case typeArray:
			a := array(c)
			for _, lo := range a.all() {
				res = append(res, key|uint64(lo))
			}
		case typeBitmap:
			b := bitmap(c)
			out := b.all()
			for _, x := range out {
				res = append(res, key|uint64(x))
			}
		}
	}
	return res
}

func (ra *Bitmap) Each(f func(value uint64)) {
	if ra == nil {
		return
	}
	ra.EachOk(func(value uint64) bool {
		f(value)
		return true
	})
}

func (ra *Bitmap) EachOk(f func(value uint64) bool) (ok bool) {
	if ra == nil {
		return
	}
	N := ra.keys.numKeys()
	for i := 0; i < N; i++ {
		key := ra.keys.key(i)
		off := ra.keys.val(i)
		c := ra.getContainer(off)

		switch c[indexType] {
		case typeArray:
			a := array(c)
			for _, lo := range a.all() {
				if !f(key | uint64(lo)) {
					return
				}
			}
		case typeBitmap:
			b := bitmap(c)
			out := b.all()
			for _, x := range out {
				if !f(key | uint64(x)) {
					return
				}
			}
		}
	}
	return true
}

func (ra *Bitmap) String() string {
	var b strings.Builder
	b.WriteRune('\n')

	var usedSize, card int
	usedSize += 4 * (ra.keys.numKeys())
	for i := 0; i < ra.keys.numKeys(); i++ {
		k := ra.keys.key(i)
		v := ra.keys.val(i)
		c := ra.getContainer(v)

		sz := c[indexSize]
		usedSize += int(sz)
		card += getCardinality(c)

		b.WriteString(fmt.Sprintf(
			"[%03d] Key: %#8x. Offset: %7d. Size: %4d. Type: %d. Card: %6d. Uint16/Uid: %.2f\n",
			i, k, v, sz, c[indexType], getCardinality(c), float64(sz)/float64(getCardinality(c))))
	}
	b.WriteString(fmt.Sprintf("Number of containers: %d. Cardinality: %d\n",
		ra.keys.numKeys(), card))

	amp := float64(len(ra.data)-usedSize) / float64(usedSize)
	b.WriteString(fmt.Sprintf(
		"Size in Uint16s. Used: %d. Total: %d. Space Amplification: %.2f%%. Moved: %.2fx\n",
		usedSize, len(ra.data), amp*100.0, float64(ra.memMoved)/float64(usedSize)))

	b.WriteString(fmt.Sprintf("Used Uint16/Uid: %.2f. Total Uint16/Uid: %.2f",
		float64(usedSize)/float64(card), float64(len(ra.data))/float64(card)))

	return b.String()
}

const fwd int = 0x01
const rev int = 0x02

func (ra *Bitmap) Minimum() uint64 { return ra.extreme(fwd) }
func (ra *Bitmap) Maximum() uint64 { return ra.extreme(rev) }

func (ra *Bitmap) Debug(x uint64) string {
	var b strings.Builder
	hi := x & mask
	off, found := ra.keys.getValue(hi)
	if !found {
		b.WriteString(fmt.Sprintf("Unable to find the container for x: %#x\n", hi))
		b.WriteString(ra.String())
	}
	c := ra.getContainer(off)
	lo := uint16(x)

	b.WriteString(fmt.Sprintf("x: %#x lo: %#x. offset: %d\n", x, lo, off))

	switch c[indexType] {
	case typeArray:
	case typeBitmap:
		idx := lo / 16
		pos := lo % 16
		b.WriteString(fmt.Sprintf("At idx: %d. Pos: %d val: %#b\n", idx, pos, c[startIdx+idx]))
	}
	return b.String()
}

func (ra *Bitmap) extreme(dir int) uint64 {
	N := ra.keys.numKeys()
	if N == 0 {
		return 0
	}

	var k uint64
	var c []uint16

	if dir == fwd {
		for i := 0; i < N; i++ {
			offset := ra.keys.val(i)
			c = ra.getContainer(offset)
			if getCardinality(c) > 0 {
				k = ra.keys.key(i)
				break
			}
		}
	} else {
		for i := N - 1; i >= 0; i-- {
			offset := ra.keys.val(i)
			c = ra.getContainer(offset)
			if getCardinality(c) > 0 {
				k = ra.keys.key(i)
				break
			}
		}
	}

	switch c[indexType] {
	case typeArray:
		a := array(c)
		if dir == fwd {
			return k | uint64(a.minimum())
		}
		return k | uint64(a.maximum())
	case typeBitmap:
		b := bitmap(c)
		if dir == fwd {
			return k | uint64(b.minimum())
		}
		return k | uint64(b.maximum())
	default:
		panic("We don't support this type of container")
	}
}

func (ra *Bitmap) And(bm *Bitmap) {
	if bm == nil {
		ra.Reset()
		return
	}

	a, b := ra, bm
	ai, an := 0, a.keys.numKeys()
	bi, bn := 0, b.keys.numKeys()

	for ai < an && bi < bn {
		ak := a.keys.key(ai)
		bk := b.keys.key(bi)
		if ak == bk {
			off := a.keys.val(ai)
			ac := a.getContainer(off)

			off = b.keys.val(bi)
			bc := b.getContainer(off)

			// do the intersection
			// TODO: See if we can do containerAnd operation in-place.
			c := containerAnd(ac, bc)

			// create a new container and update the key offset to this container.
			offset := a.newContainer(uint16(len(c)))
			copy(a.data[offset:], c)
			a.setKey(ak, offset)
			ai++
			bi++
		} else if ak < bk {
			off := a.keys.val(ai)
			zeroOutContainer(a.getContainer(off))
			ai++
		} else {
			bi++
		}
	}
	for ai < an {
		off := a.keys.val(ai)
		zeroOutContainer(a.getContainer(off))
		ai++
	}
}

func And(a, b *Bitmap) *Bitmap {
	ai, an := 0, a.keys.numKeys()
	bi, bn := 0, b.keys.numKeys()

	res := NewBitmap()
	for ai < an && bi < bn {
		ak := a.keys.key(ai)
		bk := a.keys.key(bi)
		if ak == bk {
			// Do the intersection.
			off := a.keys.val(ai)
			ac := a.getContainer(off)

			off = b.keys.val(bi)
			bc := b.getContainer(off)

			outc := containerAnd(ac, bc)
			if getCardinality(outc) > 0 {
				offset := res.newContainer(uint16(len(outc)))
				copy(res.data[offset:], outc)
				res.setKey(ak, offset)
			}
			ai++
			bi++
		} else if ak < bk {
			ai++
		} else {
			bi++
		}
	}
	return res
}

func (ra *Bitmap) AndNot(bm *Bitmap) {
	if bm == nil {
		return
	}
	a, b := ra, bm
	var ai, bi int

	buf := make([]uint16, maxContainerSize)
	for ai < a.keys.numKeys() && bi < b.keys.numKeys() {
		ak := a.keys.key(ai)
		bk := b.keys.key(bi)
		if ak == bk {
			off := a.keys.val(ai)
			ac := a.getContainer(off)

			off = b.keys.val(bi)
			bc := b.getContainer(off)

			// TODO: See if we can do containerAndNot operation in-place.
			c := containerAndNot(ac, bc, buf)
			// create a new container and update the key offset to this container.
			offset := a.newContainer(uint16(len(c)))
			copy(a.data[offset:], c)
			a.setKey(ak, offset)

			ai++
			bi++
			continue
		}
		if ak < bk {
			ai++
		} else {
			bi++
		}
	}
}

// TODO: Check if we want to use lazyMode
func (dst *Bitmap) Or(src *Bitmap) {
	if src == nil {
		return
	}
	dst.or(src, runInline)
}

func (dst *Bitmap) or(src *Bitmap, runMode int) {
	srcIdx, numKeys := 0, src.keys.numKeys()

	buf := make([]uint16, maxContainerSize)
	for ; srcIdx < numKeys; srcIdx++ {
		srcCont := src.getContainer(src.keys.val(srcIdx))
		if getCardinality(srcCont) == 0 {
			continue
		}

		key := src.keys.key(srcIdx)

		dstIdx := dst.keys.search(key)
		if dstIdx >= dst.keys.numKeys() || dst.keys.key(dstIdx) != key {
			// srcCont doesn't exist in dst. So, copy it over.
			offset := dst.newContainer(uint16(len(srcCont)))
			copy(dst.getContainer(offset), srcCont)
			dst.setKey(key, offset)
		} else {
			// Container exists in dst as well. Do an inline containerOr.
			offset := dst.keys.val(dstIdx)
			dstCont := dst.getContainer(offset)
			if c := containerOr(dstCont, srcCont, buf, runMode|runInline); len(c) > 0 {
				dst.copyAt(offset, c)
				dst.setKey(key, offset)
			}
		}
	}
}

func Or(a, b *Bitmap) *Bitmap {
	ai, an := 0, a.keys.numKeys()
	bi, bn := 0, b.keys.numKeys()

	buf := make([]uint16, maxContainerSize)
	res := NewBitmap()
	for ai < an && bi < bn {
		ak := a.keys.key(ai)
		ac := a.getContainer(a.keys.val(ai))

		bk := b.keys.key(bi)
		bc := b.getContainer(b.keys.val(bi))

		if ak == bk {
			// Do the union.
			outc := containerOr(ac, bc, buf, 0)
			offset := res.newContainer(uint16(len(outc)))
			copy(res.data[offset:], outc)
			res.setKey(ak, offset)
			ai++
			bi++
		} else if ak < bk {
			off := res.newContainer(uint16(len(ac)))
			copy(res.getContainer(off), ac)
			res.setKey(ak, off)
			ai++
		} else {
			off := res.newContainer(uint16(len(bc)))
			copy(res.getContainer(off), bc)
			res.setKey(bk, off)
			bi++
		}
	}
	for ai < an {
		ak := a.keys.key(ai)
		ac := a.getContainer(a.keys.val(ai))
		off := res.newContainer(uint16(len(ac)))

		copy(res.getContainer(off), ac)
		res.setKey(ak, off)
		ai++
	}
	for bi < bn {
		bk := b.keys.key(bi)
		bc := b.getContainer(b.keys.val(bi))
		off := res.newContainer(uint16(len(bc)))

		copy(res.getContainer(off), bc)
		res.setKey(bk, off)
		bi++
	}
	return res
}

func (ra *Bitmap) Rank(x uint64) int {
	key := x & mask
	offset, has := ra.keys.getValue(key)
	if !has {
		return -1
	}
	c := ra.getContainer(offset)
	y := uint16(x)

	// Find the rank within the container
	var rank int
	switch c[indexType] {
	case typeArray:
		rank = array(c).rank(y)
	case typeBitmap:
		rank = bitmap(c).rank(y)
	}
	if rank < 0 {
		return -1
	}

	// Add up cardinalities of all the containers on the left of container containing x.
	n := ra.keys.numKeys()
	for i := 0; i < n; i++ {
		if ra.keys.key(i) == key {
			break
		}
		cont := ra.getContainer(ra.keys.val(i))
		rank += getCardinality(cont)
	}
	return rank
}

func (ra *Bitmap) Cleanup() {
	type interval struct {
		start uint64
		end   uint64
	}

	// Find the ranges that needs to be removed in the key space and the container space. Also,
	// start the iteration from idx = 1 because we never remove the 0 key.
	var keyIntervals, contIntervals []interval
	for idx := 1; idx < ra.keys.numKeys(); idx++ {
		off := ra.keys.val(idx)
		cont := ra.getContainer(off)
		if getCardinality(cont) == 0 {
			ko := uint64(keyOffset(idx))
			contIntervals = append(contIntervals, interval{off, off + uint64(cont[indexSize])})
			keyIntervals = append(keyIntervals, interval{4 * ko, 4 * (ko + 2)})
		}
	}
	if len(contIntervals) == 0 {
		return
	}

	merge := func(intervals []interval) []interval {
		assert(len(intervals) > 0)

		// Merge the ranges in order to reduce scootLeft
		merged := []interval{intervals[0]}
		for _, ir := range intervals[1:] {
			last := merged[len(merged)-1]
			if ir.start == last.end {
				last.end = ir.end
				merged[len(merged)-1] = last
				continue
			}
			merged = append(merged, ir)
		}
		return merged
	}

	// Key intervals are already sorted, but container intervals needs to be sorted because
	// they are always added in the end of the ra.data.
	sort.Slice(contIntervals, func(i, j int) bool {
		return contIntervals[i].start < contIntervals[j].start
	})

	contIntervals = merge(contIntervals)
	keyIntervals = merge(keyIntervals)

	// Cleanup the containers.
	moved := uint64(0)
	for _, ir := range contIntervals {
		assert(ir.start >= moved)
		sz := ir.end - ir.start
		ra.scootLeft(ir.start-moved, sz)
		ra.keys.updateOffsets(ir.end-moved-1, sz, false)
		moved += sz
	}

	// Cleanup the key space.
	moved = uint64(0)
	for _, ir := range keyIntervals {
		assert(ir.start >= moved)
		sz := ir.end - ir.start
		ra.scootLeft(ir.start-moved, sz)

		// sz is in number of u16s, hence number of key-value removed is sz/8.
		ra.keys.setNumKeys(ra.keys.numKeys() - int(sz/8))
		ra.keys.setNodeSize(ra.keys.size() - int(sz))
		ra.keys = ra.keys[:len(ra.keys)-int(sz/4)]
		ra.keys.updateOffsets(ir.end-moved-1, sz, false)
		moved += sz
	}
}

func FastAnd(bitmaps ...*Bitmap) *Bitmap {
	if len(bitmaps) == 0 {
		return NewBitmap()
	}
	b := bitmaps[0]
	for _, bm := range bitmaps[1:] {
		b.And(bm)
	}
	b.Cleanup()
	return b
}

// FastParOr would group up bitmaps and call FastOr on them concurrently. It
// would then merge the groups into final Bitmap. This approach is simpler and
// faster than operating at a container level, because we can't operate on array
// containers belonging to the same Bitmap concurrently because array containers
// can expand, leaving no clear boundaries.
//
// If FastParOr is called with numGo=1, it just calls FastOr.
//
// Experiments with numGo=4 shows that FastParOr would be 2x the speed of
// FastOr, but 4x the memory usage, even under 50% CPU usage. So, use wisely.
func FastParOr(numGo int, bitmaps ...*Bitmap) *Bitmap {
	if numGo == 1 {
		return FastOr(bitmaps...)
	}
	width := max(len(bitmaps)/numGo, 3)

	var wg sync.WaitGroup
	var res []*Bitmap
	for start := 0; start < len(bitmaps); start += width {
		end := min(start+width, len(bitmaps))
		res = append(res, nil) // Make space for result.
		wg.Add(1)

		go func(start, end int) {
			idx := start / width
			res[idx] = FastOr(bitmaps[start:end]...)
			wg.Done()
		}(start, end)
	}
	wg.Wait()
	return FastOr(res...)
}

// FastOr would merge given Bitmaps into one Bitmap. This is faster than
// doing an OR over the bitmaps iteratively.
func FastOr(bitmaps ...*Bitmap) *Bitmap {
	if len(bitmaps) == 0 {
		return NewBitmap()
	}
	if len(bitmaps) == 1 {
		return bitmaps[0]
	}

	// We first figure out the container distribution across the bitmaps. We do
	// that by looking at the key of the container, and the cardinality. We
	// assume the worst-case scenario where the union would result in a
	// cardinality (per container) of the sum of cardinalities of each of the
	// corresponding containers in other bitmaps.
	containers := make(map[uint64]int)
	for _, b := range bitmaps {
		for i := 0; i < b.keys.numKeys(); i++ {
			offset := b.keys.val(i)
			cont := b.getContainer(offset)
			card := getCardinality(cont)
			containers[b.keys.key(i)] += card
		}
	}

	// We use the above information to pre-generate the destination Bitmap and
	// allocate container sizes based on the calculated cardinalities.
	// var sz int
	dst := NewBitmap()
	// First create the keys. We do this as a separate step, because keys are
	// the left most portion of the data array. Adding space there requires
	// moving a lot of pieces.
	for key, card := range containers {
		if card > 0 {
			dst.setKey(key, 0)
		}
	}

	// Then create the bitmap containers.
	for key, card := range containers {
		if card >= 4096 {
			offset := dst.newContainer(maxContainerSize)
			c := dst.getContainer(offset)
			c[indexSize] = maxContainerSize
			c[indexType] = typeBitmap
			dst.setKey(key, offset)
		}
	}

	// Create the array containers at the end. This allows them to expand
	// without having to move a lot of memory.
	for key, card := range containers {
		// Ensure this condition exactly maps up with above.
		if card < 4096 && card > 0 {
			if card < minContainerSize {
				card = minContainerSize
			}
			offset := dst.newContainer(uint16(card))
			c := dst.getContainer(offset)
			c[indexSize] = uint16(card)
			c[indexType] = typeArray
			dst.setKey(key, offset)
		}
	}

	// dst Bitmap is ready to be ORed with the given Bitmaps.
	for _, b := range bitmaps {
		dst.or(b, runLazy)
	}

	for i := 0; i < dst.keys.numKeys(); i++ {
		offset := dst.keys.val(i)
		c := dst.getContainer(offset)
		if getCardinality(c) == invalidCardinality {
			calculateAndSetCardinality(c)
		}
	}

	return dst
}

// Split splits the bitmap based on maxSz and the externalSize function. It splits the bitmap
// such that size of each split bitmap + external size corresponding to its elements approximately
// equal to maxSz (it can be greater than maxSz sometimes). The splits are returned in sorted order.
// externalSize is a function that should return the external size corresponding to elements in
// range [start, end]. External size is used to calculate the split boundaries.
func (bm *Bitmap) Split(externalSize func(start, end uint64) uint64, maxSz uint64) []*Bitmap {
	splitFurther := func(b *Bitmap) []*Bitmap {
		itr := b.NewIterator()
		newBm := NewBitmap()
		var sz uint64
		var bms []*Bitmap
		for id := itr.Next(); id != 0; id = itr.Next() {
			sz += externalSize(id, id)
			newBm.Set(id)
			if sz >= maxSz {
				bms = append(bms, newBm)
				newBm = NewBitmap()
				sz = 0
			}
		}

		if !newBm.IsEmpty() {
			bms = append(bms, newBm)
		}
		return bms
	}

	create := func(keyToOffset map[uint64]uint64, totalSz uint64) []*Bitmap {
		var keys []uint64
		for key := range keyToOffset {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		newBm := NewBitmap()

		// First set all the keys.
		var containerSz uint64
		for _, key := range keys {
			newBm.setKey(key, 0)

			// Calculate the size of the containers.
			cont := bm.getContainer(keyToOffset[key])
			containerSz += uint64(len(cont))
		}
		// Allocate enough space to hold all the containers.
		beforeSize := len(newBm.data)
		newBm.fastExpand(containerSz)
		newBm.data = newBm.data[:beforeSize]

		// Now, we can populate the containers. For that, we first expand the
		// bitmap. Calculate the total size we need to allocate all these containers.
		for _, key := range keys {
			cont := bm.getContainer(keyToOffset[key])
			off := newBm.newContainer(uint16(len(cont)))
			copy(newBm.data[off:], cont)

			newBm.setKey(key, off)
		}

		if newBm.GetCardinality() == 0 {
			return nil
		}

		if totalSz > maxSz {
			return splitFurther(newBm)
		}

		return []*Bitmap{newBm}
	}

	var splits []*Bitmap

	containerMap := make(map[uint64]uint64)
	var totalSz uint64 // size of containers plus the external size of the container

	for i := 0; i < bm.keys.numKeys(); i++ {
		key := bm.keys.key(i)
		off := bm.keys.val(i)
		cont := bm.getContainer(off)

		start, end := key, addUint64(key, 1<<16-1)
		sz := externalSize(start, end) + 2*uint64(cont[indexSize]) // Converting to bytes.

		// We can probably append more containers in the same bucket.
		if totalSz+sz < maxSz || len(containerMap) == 0 {
			// Include this container in the container map.
			containerMap[key] = off
			totalSz += sz
			continue
		}

		// We have reached the maxSz limit. Hence, create a split.
		splits = append(splits, create(containerMap, totalSz)...)

		containerMap = make(map[uint64]uint64)
		containerMap[key] = off
		totalSz = sz
	}
	if len(containerMap) > 0 {
		splits = append(splits, create(containerMap, totalSz)...)
	}

	return splits
}
