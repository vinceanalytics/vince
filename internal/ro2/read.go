package ro2

import (
	"math/bits"

	"github.com/gernest/roaring"
)

type Reader interface {
	Seek(key uint64) bool
	Valid() bool
	Next() bool
	Value() (uint64, *roaring.Container)
	Max() uint64
	OffsetRange(offset, start, end uint64) *Bitmap
	ApplyFilter(key uint64, filter roaring.BitmapFilter) (err error)
}

func ReadSum(ra Reader, match *Bitmap) (count int32, total int64) {
	fs := roaring.NewBitmapBSICountFilter(match)
	ra.ApplyFilter(0, fs)
	return fs.Total()
}

func ReadTrue(ra Reader, shard uint64, match *Bitmap) *Bitmap {
	m := Row(ra, shard, trueRowID)
	return m.Intersect(match)
}

func ReadFalse(ra Reader, shard uint64, match *Bitmap) *Bitmap {
	m := Row(ra, shard, falseRowID)
	return m.Intersect(match)
}

func ReadMutex(ra Reader, shard uint64, filterBitmap *Bitmap, match func(row uint64, ra *Bitmap)) {
	filter := make([]*roaring.Container, 1<<shardVsContainerExponent)
	filterIterator, _ := filterBitmap.Containers.Iterator(0)
	// So let's get these all with a nice convenient 0 offset...
	for filterIterator.Next() {
		k, c := filterIterator.Value()
		if c.N() == 0 {
			continue
		}
		filter[k%(1<<shardVsContainerExponent)] = c
	}

	prevRow := ^uint64(0)
	seenThisRow := false
	resultContainerKey := shard * rowWidth
	for ra.Seek(0); ra.Valid(); ra.Next() {
		k, c := ra.Value()
		row := k >> shardVsContainerExponent
		if row == prevRow {
			if seenThisRow {
				continue
			}
		} else {
			seenThisRow = false
			prevRow = row
		}
		if roaring.IntersectionAny(c, filter[k%(1<<shardVsContainerExponent)]) {
			ro := NewBitmap()
			ro.Containers.Put(resultContainerKey, roaring.Intersect(c, filter[k%(1<<shardVsContainerExponent)]))
			match(row, ro)
			seenThisRow = true
		}
	}
}

func ReadDistinctBSI(ra Reader, shard uint64, filterBitmap *Bitmap) *roaring.Bitmap {
	existsBitmap := ra.OffsetRange(shardWidth*shard, shardWidth*0, shardWidth*1)
	if filterBitmap != nil {
		existsBitmap = existsBitmap.Intersect(filterBitmap)
	}
	signBitmap := ra.OffsetRange(shardWidth*shard, shardWidth*1, shardWidth*2)

	depth := ra.Max() / shardWidth
	dataBitmaps := make([]*roaring.Bitmap, depth)
	for i := uint64(0); i < depth; i++ {
		dataBitmaps[i] = ra.OffsetRange(shardWidth*shard, shardWidth*(i+2), shardWidth*(i+3))
	}
	stashWords := make([]uint64, 1024*(depth+2))
	bitStashes := make([][]uint64, depth)
	for i := uint64(0); i < depth; i++ {
		start := i * 1024
		last := start + 1024
		bitStashes[i] = stashWords[start:last]
		i++
	}
	stashOffset := depth * 1024
	existStash := stashWords[stashOffset : stashOffset+1024]
	signStash := stashWords[stashOffset+1024 : stashOffset+2048]
	dataBits := make([][]uint64, depth)

	posValues := make([]uint64, 0, 64)
	negValues := make([]uint64, 0, 64)

	posBitmap := roaring.NewFileBitmap()
	negBitmap := roaring.NewFileBitmap()

	existIterator, _ := existsBitmap.Containers.Iterator(0)
	for existIterator.Next() {
		key, value := existIterator.Value()
		if value.N() == 0 {
			continue
		}
		exists := value.AsBitmap(existStash)
		sign := signBitmap.Containers.Get(key).AsBitmap(signStash)
		for i := uint64(0); i < depth; i++ {
			dataBits[i] = dataBitmaps[i].Containers.Get(key).AsBitmap(bitStashes[i])
		}
		for idx, word := range exists {
			// mask holds a mask we can test the other words against.
			mask := uint64(1)
			for word != 0 {
				shift := uint(bits.TrailingZeros64(word))
				// we shift one *more* than that, to move the
				// actual one bit off.
				word >>= shift + 1
				mask <<= shift
				value := int64(0)
				for b := uint64(0); b < depth; b++ {
					if dataBits[b][idx]&mask != 0 {
						value += (1 << b)
					}
				}
				if sign[idx]&mask != 0 {
					value *= -1
				}
				if value < 0 {
					negValues = append(negValues, uint64(-value))
				} else {
					posValues = append(posValues, uint64(value))
				}
				// and now we processed that bit, so we move the mask over one.
				mask <<= 1
			}
			if len(negValues) > 0 {
				_, _ = negBitmap.AddN(negValues...)
				negValues = negValues[:0]
			}
			if len(posValues) > 0 {
				_, _ = posBitmap.AddN(posValues...)
				posValues = posValues[:0]
			}
		}
	}
	return posBitmap.Union(negBitmap)
}

type OffsetRanger interface {
	OffsetRange(offset, start, end uint64) *Bitmap
}

func Row(ra OffsetRanger, shard, rowID uint64) *Bitmap {
	return ra.OffsetRange(shardWidth*shard, shardWidth*rowID, shardWidth*(rowID+1))
}

func Existence(tx OffsetRanger, shard uint64) *Bitmap {
	return Row(tx, shard, bsiExistsBit)
}
