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
	"math/rand"
	"runtime"
	"testing"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
)

// go test -bench BenchmarkMemoryUsage -run -
func BenchmarkMemoryUsage(b *testing.B) {
	b.StopTimer()
	bitmaps := make([]*Bitmap, 0, 10)

	incr := uint64(1 << 16)
	max := uint64(1<<32 - 1)
	for x := 0; x < 10; x++ {
		rb := NewBitmap()

		var i uint64
		for i = 0; i <= max-incr; i += incr {
			rb.Set(i)
		}

		bitmaps = append(bitmaps, rb)
	}

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	b.Logf("HeapInUse: %d, HeapObjects: %d", stats.HeapInuse, stats.HeapObjects)
	b.StartTimer()
}

// go test -bench BenchmarkIntersection -run -
func BenchmarkIntersectionRoaring(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	s1 := NewBitmap()
	sz := int64(150000)
	initsize := 65000
	for i := 0; i < initsize; i++ {
		s1.Set(uint64(r.Int63n(sz)))
	}

	s2 := NewBitmap()
	sz = int64(100000000)
	initsize = 65000
	for i := 0; i < initsize; i++ {
		s2.Set(uint64(r.Int63n((sz))))
	}
	b.StartTimer()

	card := 0
	for j := 0; j < b.N; j++ {
		s3 := And(s1, s2)
		card = card + s3.GetCardinality()
	}
	b.Logf("card: %d\n", card)
}

// go test -bench BenchmarkSet -run -
func BenchmarkSetRoaring(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := int64(1000000)
	s := NewBitmap()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Set(uint64(r.Int63n(sz)))
	}
}

func BenchmarkMerge10K(b *testing.B) {
	var bitmaps []*Bitmap
	for i := 0; i < 10000; i++ {
		bm := NewBitmap()
		for j := 0; j < 1000; j++ {
			x := rand.Uint64() % 1e8 // 10M.
			bm.Set(x)
		}
		bitmaps = append(bitmaps, bm)
	}

	second := func() *Bitmap {
		var res []*Bitmap
		for i := 0; i < 100; i += 1 {
			input := bitmaps[100*i : 100*i+100]
			out := FastOr(input...)
			res = append(res, out)
		}
		return FastOr(res...)
	}

	out := FastOr(bitmaps...)
	b.Logf("Out: %s\n", out)
	out2 := second()
	if out2.GetCardinality() != out.GetCardinality() {
		panic("Don't match")
	}
	out3 := FastParOr(8, bitmaps...)
	if out3.GetCardinality() != out.GetCardinality() {
		panic("Don't match")
	}
	b.Logf("card2: %d card3: %d", out2.GetCardinality(), out3.GetCardinality())

	b.Run("fastor", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FastOr(bitmaps...)
		}
	})

	b.Run("fastor-groups", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = second()
		}
	})
	b.Run("fastparor", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FastParOr(4, bitmaps...)
		}
	})
}

func BenchmarkRemoveRange(b *testing.B) {
	bm := NewBitmap()
	N := uint64(1e5)
	for i := uint64(0); i < N; i++ {
		bm.Set(uint64(i))
	}

	bench := func(b *testing.B, factor uint64) {
		sz := uint64(N / factor)
		cnt := uint64(N / sz)
		for j := 0; j < b.N; j++ {
			b.StopTimer()
			bm2 := bm.Clone()
			b.StartTimer()
			for i := uint64(0); i < cnt; i++ {
				bm2.RemoveRange(i*sz, (i+1)*sz)
			}
		}
	}
	b.Run("N/2", func(b *testing.B) {
		bench(b, 2)
	})
	b.Run("N/4", func(b *testing.B) {
		bench(b, 4)
	})
	b.Run("N/16", func(b *testing.B) {
		bench(b, 16)
	})
	b.Run("N/256", func(b *testing.B) {
		bench(b, 256)
	})
}

func BenchmarkRemoveRangeRoaring64(b *testing.B) {
	bm := roaring64.NewBitmap()
	N := uint64(1e5)
	for i := uint64(0); i < N; i++ {
		bm.Add(uint64(i))
	}

	bench := func(b *testing.B, factor uint64) {
		sz := uint64(N / factor)
		cnt := uint64(N / sz)
		for j := 0; j < b.N; j++ {
			b.StopTimer()
			bm2 := bm.Clone()
			b.StartTimer()
			for i := uint64(0); i < cnt; i++ {
				bm2.RemoveRange(i*sz, (i+1)*sz)
			}
		}
	}
	b.Run("N/2", func(b *testing.B) {
		bench(b, 2)
	})
	b.Run("N/4", func(b *testing.B) {
		bench(b, 4)
	})
	b.Run("N/16", func(b *testing.B) {
		bench(b, 16)
	})
	b.Run("N/256", func(b *testing.B) {
		bench(b, 256)
	})
}

func BenchmarkSelectSroar(b *testing.B) {
	bm := NewBitmap()
	N := uint64(1e5)
	for i := uint64(0); i < N; i++ {
		bm.Set(uint64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < N; j++ {
			bm.Select(j)
		}
	}
}

func BenchmarkSelectRoaring64(b *testing.B) {
	bm := roaring64.NewBitmap()
	N := uint64(1e5)
	for i := uint64(0); i < N; i++ {
		bm.Add(uint64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := uint64(0); j < N; j++ {
			bm.Select(j)
		}
	}
}
