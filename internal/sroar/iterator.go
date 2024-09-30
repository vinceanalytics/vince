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
	"math/bits"
)

type Iterator struct {
	bm *Bitmap

	keys   []uint64
	keyIdx int

	contIdx int

	bitmapIdx int
	bitset    uint16
}

func (bm *Bitmap) NewRangeIterators(numRanges int) []*Iterator {
	keyn := bm.keys.numKeys()
	iters := make([]*Iterator, numRanges)
	width := keyn / numRanges
	rem := keyn % numRanges
	cnt := 0

	// This loop distributes the key equally to the ranges. For example: If numRanges = 3
	// and keyn = 8 then it will be distributes as [3, 3, 2]
	for i := 0; i < numRanges; i++ {
		iters[i] = bm.NewIterator()
		n := width
		if i < rem {
			n = width + 1
		}
		iters[i].keys = iters[i].keys[cnt : cnt+2*n]
		cnt = cnt + 2*n
	}
	return iters
}

func (bm *Bitmap) NewIterator() *Iterator {
	return &Iterator{
		bm:        bm,
		keys:      bm.keys[indexNodeStart : indexNodeStart+bm.keys.numKeys()*2],
		keyIdx:    0,
		contIdx:   -1,
		bitmapIdx: -1,
	}
}

func (it *Iterator) Next() uint64 {
	if len(it.keys) == 0 {
		return 0
	}

	key := it.keys[it.keyIdx]
	off := it.keys[it.keyIdx+1]
	cont := it.bm.getContainer(off)
	card := getCardinality(cont)

	// Loop until we find a container on which next operation is possible. When such a container
	// is found, reset the variables responsible for container iteration.
	for card == 0 || it.contIdx+1 >= card {
		if it.keyIdx+2 >= len(it.keys) {
			return 0
		}
		// jump by 2 because key is followed by a value
		it.keyIdx += 2
		it.contIdx = -1
		it.bitmapIdx = -1
		it.bitset = 0
		key = it.keys[it.keyIdx]
		off = it.keys[it.keyIdx+1]
		cont = it.bm.getContainer(off)
		card = getCardinality(cont)
	}

	//  The above loop assures that we can do next in this container.
	it.contIdx++
	switch cont[indexType] {
	case typeArray:
		return key | uint64(cont[int(startIdx)+it.contIdx])
	case typeBitmap:
		// A bitmap container is an array of uint16s.
		// If the container is bitmap, go to the index which has a non-zero value.
		for it.bitset == 0 && it.bitmapIdx+1 < len(cont[startIdx:]) {
			it.bitmapIdx++
			it.bitset = cont[int(startIdx)+it.bitmapIdx]
		}
		assert(it.bitset > 0)

		// msbIdx is the index of most-significant bit. In this iteration we choose this set bit
		// and make it zero.
		msbIdx := uint16(bits.LeadingZeros16(it.bitset))
		msb := 1 << (16 - msbIdx - 1)
		it.bitset ^= uint16(msb)
		return key | uint64(it.bitmapIdx*16+int(msbIdx))
	}
	return 0
}

type ManyItr struct {
	index int
	arr   []uint64
}

// TODO: See if this is needed, we should remove this
func (r *Bitmap) ManyIterator() *ManyItr {
	return &ManyItr{
		arr: r.ToArray(),
	}

}

func (itr *ManyItr) NextMany(buf []uint64) int {
	count := 0
	for i := 0; i < len(buf); i++ {
		if itr.index == len(itr.arr) {
			break
		}
		buf[i] = itr.arr[itr.index]
		itr.index++
		count++
	}
	return count
}
