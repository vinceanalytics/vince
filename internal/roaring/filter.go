package roaring

import (
	"math/bits"
)

const (
	exponent                 = 20
	shardWidth               = 1 << exponent
	rowExponent              = (exponent - 16)  // for instance, 20-16 = 4
	rowWidth                 = 1 << rowExponent // containers per row, for instance 1<<4 = 16
	keyMask                  = (rowWidth - 1)   // a mask for offset within the row
	shardVsContainerExponent = 4
)

func (ra *Bitmap) ExtractMutex(match *Bitmap, f func(row uint64, columns *Bitmap) error) error {
	if ra == nil {
		return nil
	}
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
			err := f(row, toRows(ex))
			if err != nil {
				return err
			}
			seenThisRow = true
		}
	}
	return nil
}

func highbits(v uint64) uint64 { return v >> 16 }

func toRows(ac []uint16) *Bitmap {
	res := NewBitmap()
	offs := res.newContainer(uint16(len(ac)))
	copy(res.getContainer(offs), ac)
	res.setKey(0, offs)
	return res
}

func (ra *Bitmap) BSI(id uint64, svalue int64) {
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

const (
	// BSI bits used to check existence & sign.
	bsiExistsBit = 0
	bsiSignBit   = 1
	bsiOffsetBit = 2
)

func (ra *Bitmap) ExtractBSI(shard uint64, match *Bitmap, f func(id uint64, value int64)) {
	if ra == nil {
		return
	}
	exists := ra.Row(shard, bsiExistsBit)
	exists.And(match)
	if exists.IsEmpty() {
		return
	}
	data := make(map[uint64]uint64, exists.GetCardinality())
	mergeBits(exists, 0, data)
	sign := ra.Row(shard, bsiSignBit)
	sign.And(exists)
	mergeBits(sign, 1<<63, data)
	bitDepth := ra.depth()
	for i := uint64(0); i < bitDepth; i++ {
		bits := ra.Row(shard, bsiOffsetBit+uint64(i))
		bits.And(exists)
		mergeBits(bits, 1<<i, data)
	}
	for columnID, val := range data {
		// Convert to two's complement and add base back to value.
		val = uint64((2*(int64(val)>>63) + 1) * int64(val&^(1<<63)))
		f(columnID, int64(val))
	}
}

func (ra *Bitmap) BSISum(shard uint64, match *Bitmap) (sum int64) {
	exists := ra.Row(shard, bsiExistsBit)
	exists.And(match)
	if exists.IsEmpty() {
		return
	}
	sign := ra.Row(shard, bsiSignBit)
	sign.And(exists)
	bitDepth := ra.depth()
	exists.AndNot(sign)
	var (
		psum, nsum uint64
	)
	for i := uint64(0); i < bitDepth; i++ {
		bits := ra.Row(shard, bsiOffsetBit+uint64(i))
		pcount := AndCardinality(exists, bits)
		ncount := AndCardinality(sign, bits)
		psum += pcount << uint(i)
		nsum += ncount << uint(i)
	}
	return int64(psum) - int64(nsum)
}

func mergeBits(bits *Bitmap, mask uint64, out map[uint64]uint64) {
	bits.Each(func(value uint64) {
		out[value] |= mask
	})
}

func (ra *Bitmap) depth() uint64 {
	return ra.Maximum() / shardWidth
}

const (
	// Row ids used for boolean fields.
	falseRowID = uint64(0)
	trueRowID  = uint64(1)

	falseRowOffset = 0 * shardWidth // fragment row 0
	trueRowOffset  = 1 * shardWidth // fragment row 1
)

func (ra *Bitmap) Bool(id uint64, value bool) {
	fragmentColumn := id % shardWidth
	if value {
		ra.Set(trueRowOffset + fragmentColumn)
	} else {
		ra.Set(falseRowOffset + fragmentColumn)
	}
}

func (ra *Bitmap) True(shard uint64, match *Bitmap) *Bitmap {
	m := ra.Row(shard, trueRowID)
	m.And(match)
	return m
}

func (ra *Bitmap) False(shard uint64, match *Bitmap) *Bitmap {
	m := ra.Row(shard, falseRowID)
	m.And(match)
	return m
}

func (ra *Bitmap) Mutex(id uint64, value uint64) {
	ra.Set(value*shardWidth + (id % shardWidth))
}

func (ra *Bitmap) Row(shard, rowID uint64) *Bitmap {
	if ra == nil {
		return NewBitmap()
	}
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
