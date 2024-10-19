package roaring

const (
	exponent                 = 20
	shardWidth               = 1 << exponent
	rowExponent              = (exponent - 16)  // for instance, 20-16 = 4
	rowWidth                 = 1 << rowExponent // containers per row, for instance 1<<4 = 16
	keyMask                  = (rowWidth - 1)   // a mask for offset within the row
	shardVsContainerExponent = 4
)

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
