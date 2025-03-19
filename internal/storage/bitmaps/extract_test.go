package bitmaps

import (
	"iter"
	"testing"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
)

func TestExtractEquality(t *testing.T) {
	CheckMutexAll(t, mutex)
	CheckMutexSubset(t, mutex)
	CheckMutexSubsetWithEmpty(t, mutex)
}

func mutex(s iter.Seq2[uint64, uint64]) Iter {
	ra := roaring.NewBitmap()
	for id, v := range s {
		Equality(ra, id, v)
	}
	it, _ := ra.Containers.Iterator(0)
	return &Wrap{ContainerIterator: it}
}

func TestCompare(t *testing.T) {
	CheckCompareEq(t, build)
	CheckCompareGT(t, build)
	CheckCompareGE(t, build)
	CheckCompareLT(t, build)
	CheckCompareLE(t, build)
	CheckCompareRangeSimple(t, build)
}

func build(s iter.Seq2[uint64, int64]) (depth uint64, offset OffsetRanger) {
	ra := roaring.NewBitmap()
	for id, v := range s {
		BitSliced(ra, id, v)
	}
	return ra.Max() / shardwidth.ShardWidth, ra
}
