package roaring

import (
	"math/rand"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSetAndGet(t *testing.T) {

	bsi := NewBSI(999, 0)
	require.NotNil(t, bsi.bA)
	require.Equal(t, 10, len(bsi.bA))

	bsi.SetValue(1, 8)
	gv, ok := bsi.GetValue(1)
	require.True(t, ok)
	require.Equal(t, int64(8), gv)
}

func setup() *BSI {

	bsi := NewBSI(100, 0)
	// Setup values
	for i := 0; i < int(bsi.MaxValue); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
	return bsi
}

func setupNegativeBoundary() *BSI {

	bsi := NewBSI(5, -5)
	// Setup values
	for i := int(bsi.MinValue); i <= int(bsi.MaxValue); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
	return bsi
}

func setupAllNegative() *BSI {
	bsi := NewBSI(-1, -100)
	// Setup values
	for i := int(bsi.MinValue); i <= int(bsi.MaxValue); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
	return bsi
}

func setupAutoSizeNegativeBoundary() *BSI {
	bsi := NewDefaultBSI()
	// Setup values
	for i := int(-5); i <= int(5); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}
	return bsi
}

func setupRandom() *BSI {
	bsi := NewBSI(99, -1)
	rg := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Setup values
	for i := 0; bsi.GetExistenceBitmap().GetCardinality() < 100; {
		rv := rg.Int63n(bsi.MaxValue) - 50
		_, ok := bsi.GetValue(uint64(i))
		if ok {
			continue
		}
		bsi.SetValue(uint64(i), rv)
		i++
	}
	batch := make([]uint64, 100)
	iter := bsi.GetExistenceBitmap().ManyIterator()
	iter.NextMany(batch)
	var min, max int64
	min = Max64BitSigned
	max = Min64BitSigned
	for i := 0; i < len(batch); i++ {
		v, _ := bsi.GetValue(batch[i])
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	bsi.MinValue = min
	bsi.MaxValue = max
	return bsi
}

func TestEQ(t *testing.T) {
	bsi := setup()
	eq := bsi.CompareValue(0, EQ, 50, 0, nil)
	require.Equal(t, 1, eq.GetCardinality())
	require.True(t, eq.Contains(50))
}

func TestLT(t *testing.T) {

	bsi := setup()
	lt := bsi.CompareValue(0, LT, 50, 0, nil)
	require.Equal(t, 50, lt.GetCardinality())
	a := lt.ToArray()
	slices.Sort(a)
	require.Less(t, a[len(a)-1], uint64(50))
}

func TestGT(t *testing.T) {

	bsi := setup()
	gt := bsi.CompareValue(0, GT, 50, 0, nil)
	require.Equal(t, 49, gt.GetCardinality())

	a := gt.ToArray()
	slices.Sort(a)
	require.Greater(t, a[0], uint64(50))
}

func TestGE(t *testing.T) {

	bsi := setup()
	ge := bsi.CompareValue(0, GE, 50, 0, nil)
	require.Equal(t, 50, ge.GetCardinality())

	a := ge.ToArray()
	slices.Sort(a)
	require.GreaterOrEqual(t, a[0], uint64(50))
}

func TestLE(t *testing.T) {

	bsi := setup()
	le := bsi.CompareValue(0, LE, 50, 0, nil)
	require.Equal(t, 51, le.GetCardinality())

	a := le.ToArray()
	slices.Sort(a)
	require.LessOrEqual(t, a[len(a)-1], uint64(50))
}

func TestRange(t *testing.T) {

	bsi := setup()
	set := bsi.CompareValue(0, RANGE, 45, 55, nil)
	require.Equal(t, 11, set.GetCardinality())

	a := set.ToArray()
	slices.Sort(a)
	require.GreaterOrEqual(t, a[0], uint64(45))
	require.LessOrEqual(t, a[len(a)-1], uint64(55))
}

func TestExists(t *testing.T) {

	bsi := NewBSI(10, 0)
	// Setup values
	for i := 1; i < int(bsi.MaxValue); i++ {
		bsi.SetValue(uint64(i), int64(i))
	}

	require.Equal(t, uint64(9), bsi.GetCardinality())
	require.False(t, bsi.ValueExists(uint64(0)))
	bsi.SetValue(uint64(0), int64(0))
	require.Equal(t, uint64(10), bsi.GetCardinality())
	require.True(t, bsi.ValueExists(uint64(0)))
}

func TestRangeAllNegative(t *testing.T) {
	bsi := setupAllNegative()
	require.Equal(t, uint64(100), bsi.GetCardinality())
	set := bsi.CompareValue(0, RANGE, -55, -45, nil)
	require.Equal(t, 11, set.GetCardinality())

	a := set.ToArray()
	for i := range a {
		val, _ := bsi.GetValue(a[i])
		require.GreaterOrEqual(t, val, int64(-55))
		require.LessOrEqual(t, val, int64(-45))
	}
}

func TestMinMaxWithRandom(t *testing.T) {
	bsi := setupRandom()
	require.Equal(t, bsi.MinValue, bsi.MinMax(0, MIN, bsi.GetExistenceBitmap()))
	require.Equal(t, bsi.MaxValue, bsi.MinMax(0, MAX, bsi.GetExistenceBitmap()))
}

func TestUint32(t *testing.T) {
	want := []uint32{1, 2, 3}
	require.Equal(t, want, toUint32Slice(toBytes(want)))
}

func TestBSI_ToBuffer(t *testing.T) {
	t.Run("empty BSI is 0 bytes", func(t *testing.T) {
		b := NewDefaultBSI()
		require.Equal(t, 0, b.GetSizeInBytes())
		require.Equal(t, []byte{}, b.ToBuffer())
	})
	t.Run("rind trip", func(t *testing.T) {
		bsi := setupRandom()

		data := bsi.ToBuffer()
		bsi2 := NewBSIFromBuffer(data)

		require.Equal(t, bsi.MinValue, bsi2.MinMax(0, MIN, bsi2.GetExistenceBitmap()))
		require.Equal(t, bsi.MaxValue, bsi2.MinMax(0, MAX, bsi2.GetExistenceBitmap()))
	})
}

func BenchmarkFromBuffer(b *testing.B) {
	bsi := setup()
	data := bsi.ToBuffer()
	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		NewBSIFromBuffer(data)
	}
}
func BenchmarkBSI(b *testing.B) {
	bsi := setup()

	b.Run("ToBuffer", func(b *testing.B) {

		for range b.N {
			bsi.ToBuffer()
		}
	})
	b.Run("ToBufferWith", func(b *testing.B) {
		off := make([]uint32, 0, 1+bsi.BitCount())
		data := make([]byte, 0, bsi.GetSizeInBytes())
		for range b.N {
			bsi.ToBufferWith(off, data)
		}
	})

}

func TestSum(t *testing.T) {

	bsi := setup()
	set := bsi.CompareValue(0, RANGE, 45, 55, nil)

	sum, count := bsi.Sum(set)
	require.Equal(t, uint64(11), count)
	require.Equal(t, int64(550), sum)
}

func TestSumWithNegative(t *testing.T) {
	bsi := setupNegativeBoundary()
	require.Equal(t, uint64(11), bsi.GetCardinality())
	sum, cnt := bsi.Sum(bsi.GetExistenceBitmap())
	require.Equal(t, uint64(11), cnt)
	require.Equal(t, int64(0), sum)
}

func TestExtract(t *testing.T) {
	bsi := setup()
	set := bsi.CompareValue(0, RANGE, 45, 55, nil)
	want := map[uint64]int64{}
	set.Each(func(value uint64) {
		want[value] = int64(value)
	})
	require.Equal(t, want, bsi.Extract(set))
}

func TestBSIOr(t *testing.T) {
	a := NewDefaultBSI()
	a.SetValue(1, 100)
	b := NewDefaultBSI()
	b.SetValue(3, 1999)
	c := a.Or(b)
	want, _ := os.ReadFile("testdata/bsi_or.txt")
	require.Equal(t, string(want), c.String())
}

func BenchmarkBSIExtract(b *testing.B) {
	bsi := setup()
	set := bsi.CompareValue(0, RANGE, 45, 55, nil)

	for range b.N {
		bsi.Extract(set)
	}
}
