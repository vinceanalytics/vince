package bitmaps

import (
	"iter"
	"sync"
	"testing"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/stretchr/testify/require"
)

type Extractor struct {
	filter [1 << ShardVsContainerExponent]*roaring.Container
	data   map[uint64]uint64
}

var extractorPool = &sync.Pool{New: func() any {
	return &Extractor{data: make(map[uint64]uint64)}
}}

func NewExtractor() *Extractor {
	return extractorPool.Get().(*Extractor)
}

func (b *Extractor) Release() {
	extractorPool.Put(b)
}

func (b *Extractor) Mutex(ra Iter, shard uint64, match *roaring.Bitmap) (result [][]uint64) {
	result = make([][]uint64, match.Count())
	positions := buildPositions(match)

	filterIterator, _ := match.Containers.Iterator(0)
	for filterIterator.Next() {
		k, c := filterIterator.Value()
		if c.N() == 0 {
			continue
		}
		b.filter[k%(1<<ShardVsContainerExponent)] = c
	}

	prevRow := ^uint64(0)
	seenThisRow := false
	off := highbits(shardwidth.ShardWidth * shard)
	for ra.Seek(0); ra.Valid(); ra.Next() {
		k, c := ra.Value()
		row := k >> ShardVsContainerExponent
		if row == prevRow {
			if seenThisRow {
				continue
			}
		} else {
			seenThisRow = false
			prevRow = row
		}
		hi0 := highbits(shardwidth.ShardWidth * row)
		base := off + k - hi0
		roaring.IntersectionCallback(c, b.filter[k%(1<<ShardVsContainerExponent)], func(u uint16) {
			if !seenThisRow {
				seenThisRow = true
			}
			id := base<<16 | uint64(u)
			idx := positions[id]
			result[idx] = append(result[idx], row)
		})
	}
	clear(b.filter[:])
	return result
}

func highbits(v uint64) uint64 { return v >> 16 }

func (b *Extractor) BSI(r OffsetRanger, bitDepth, shard uint64, filter *roaring.Bitmap) (result []uint64) {
	result = make([]uint64, filter.Count())

	exists := Row(r, shard, bsiExistsBit)
	exists = exists.Intersect(filter)
	if !exists.Any() {
		return result
	}

	mergeBits(exists, 0, b.data)

	sign := Row(r, shard, bsiSignBit)
	mergeBits(sign, 1<<63, b.data)

	for i := uint64(0); i < bitDepth; i++ {
		bits := Row(r, shard, bsiOffsetBit+i)

		bits = bits.Intersect(exists)
		mergeBits(bits, 1<<i, b.data)
	}

	itr := filter.Iterator()
	itr.Seek(0)
	var i int
	for v, eof := itr.Next(); !eof; v, eof = itr.Next() {
		val, ok := b.data[v]
		if ok {
			result[i] = uint64((2*(int64(val)>>63) + 1) * int64(val&^(1<<63)))
		}
		i++
	}
	clear(b.data)
	return result
}

func mergeBits(ra *roaring.Bitmap, mask uint64, out map[uint64]uint64) {
	itr := ra.Iterator()
	itr.Seek(0)
	for v, eof := itr.Next(); !eof; v, eof = itr.Next() {
		out[v] |= mask
	}
}

type MutexSource func(iter.Seq2[uint64, uint64]) Iter

func CheckMutexAll(t *testing.T, source MutexSource) {
	t.Helper()
	type T struct {
		values []uint64
		id     uint64
	}

	samples := []T{
		{id: 1, values: []uint64{1, 2, 3, 4}},
		{id: 2, values: []uint64{3}},
		{id: 3, values: []uint64{5, 6, 7}},
	}
	it := source(func(yield func(uint64, uint64) bool) {
		for _, v := range samples {
			for _, a := range v.values {
				if !yield(v.id, a) {
					return
				}
			}
		}
	})
	ex := NewExtractor()
	defer ex.Release()
	got := ex.Mutex(it, 0, roaring.NewBitmap(1, 2, 3))
	var want [][]uint64
	for _, s := range samples {
		want = append(want, s.values)
	}
	require.Equal(t, want, got)
}

func CheckMutexSubset(t *testing.T, source MutexSource) {
	t.Helper()
	type T struct {
		values []uint64
		id     uint64
	}

	samples := []T{
		{id: 1, values: []uint64{1, 2, 3, 4}},
		{id: 2, values: []uint64{3}},
		{id: 3, values: []uint64{5, 6, 7}},
	}
	it := source(func(yield func(uint64, uint64) bool) {
		for _, v := range samples {
			for _, a := range v.values {
				if !yield(v.id, a) {
					return
				}
			}
		}
	})
	ex := NewExtractor()
	defer ex.Release()
	got := ex.Mutex(it, 0, roaring.NewBitmap(1, 3))

	require.Equal(t, [][]uint64{
		samples[0].values, samples[2].values,
	}, got)
}

func CheckMutexSubsetWithEmpty(t *testing.T, source MutexSource) {
	t.Helper()
	type T struct {
		values []uint64
		id     uint64
	}

	samples := []T{
		{id: 1, values: []uint64{1, 2, 3, 4}},
		{id: 2, values: []uint64{3}},
		{id: 3, values: []uint64{5, 6, 7}},
	}
	it := source(func(yield func(uint64, uint64) bool) {
		for _, v := range samples {
			for _, a := range v.values {
				if !yield(v.id, a) {
					return
				}
			}
		}
	})
	ex := NewExtractor()
	defer ex.Release()
	got := ex.Mutex(it, 0, roaring.NewBitmap(3, 4))

	require.Equal(t, [][]uint64{
		samples[2].values, nil,
	}, got)
}

type BSISource func(iter.Seq2[uint64, int64]) (depth uint64, offset OffsetRanger)

func CheckCompareEq(t *testing.T, source BSISource) {
	t.Helper()
	depth, bsi := source(setup())
	eq := Range(bsi, EQ, 0, depth, 50, 0)
	require.Equal(t, uint64(1), eq.Count())
	require.True(t, eq.Contains(50))
}

func CheckCompareLT(t *testing.T, source BSISource) {
	t.Helper()
	depth, bsi := source(setup())
	lt := Range(bsi, LT, 0, depth, 50, 0)
	require.Equal(t, uint64(49), lt.Count())
	it := lt.Iterator()
	it.Seek(0)
	for v, eof := it.Next(); !eof; v, eof = it.Next() {
		require.Less(t, v, uint64(50))
	}
}

func CheckCompareGT(t *testing.T, source BSISource) {
	t.Helper()
	depth, bsi := source(setup())

	gt := Range(bsi, GT, 0, depth, 50, 0)
	require.Equal(t, uint64(50), gt.Count())
	it := gt.Iterator()
	it.Seek(0)
	for v, eof := it.Next(); !eof; v, eof = it.Next() {
		require.Greater(t, v, uint64(50))
	}
}

func CheckCompareGE(t *testing.T, source BSISource) {
	t.Helper()
	depth, bsi := source(setup())
	ge := Range(bsi, GTE, 0, depth, 50, 0)
	require.Equal(t, uint64(51), ge.Count())
	it := ge.Iterator()
	it.Seek(0)
	for v, eof := it.Next(); !eof; v, eof = it.Next() {
		require.GreaterOrEqual(t, v, uint64(50))
	}
}

func CheckCompareLE(t *testing.T, source BSISource) {
	t.Helper()
	depth, bsi := source(setup())
	le := Range(bsi, LTE, 0, depth, 50, 0)
	require.Equal(t, uint64(50), le.Count())
	it := le.Iterator()
	it.Seek(0)
	for v, eof := it.Next(); !eof; v, eof = it.Next() {
		require.LessOrEqual(t, v, uint64(50))
	}
}

func CheckCompareRangeSimple(t *testing.T, source BSISource) {
	t.Helper()
	depth, bsi := source(setup())
	le := Range(bsi, BETWEEN, 0, depth, 45, 55)
	require.Equal(t, uint64(11), le.Count())
	it := le.Iterator()
	it.Seek(0)
	ex := NewExtractor()
	defer ex.Release()
	for v, eof := it.Next(); !eof; v, eof = it.Next() {
		require.GreaterOrEqual(t, v, uint64(45))
		require.LessOrEqual(t, v, uint64(55))

		result := ex.BSI(bsi, depth, 0, roaring.NewBitmap(v))
		require.Len(t, result, 1)
		require.Equal(t, result[0], v)
	}
}

func setup() iter.Seq2[uint64, int64] {
	return func(yield func(uint64, int64) bool) {
		for n := 1; n <= 100; n++ {
			if !yield(uint64(n), int64(n)) {
				return
			}
		}
	}
}

func buildPositions(match *roaring.Bitmap) (m map[uint64]int) {
	m = make(map[uint64]int)
	itr := match.Iterator()
	itr.Seek(0)
	for v, eof := itr.Next(); !eof; v, eof = itr.Next() {
		m[v] = len(m)
	}
	return
}
