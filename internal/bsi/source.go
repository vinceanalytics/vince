package bsi

import "github.com/vinceanalytics/vince/internal/roaring"

type Source interface {
	BitCount() int
	Get(i int) *roaring.Bitmap
	GetOrCreate(i int) *roaring.Bitmap
}

type Reset interface {
	Reset()
}

type Slice struct {
	a []*roaring.Bitmap
}

func (s *Slice) BitCount() int {
	if len(s.a) == 0 {
		return 0
	}
	return len(s.a) - 1
}

func (s *Slice) Reset() {
	clear(s.a)
	s.a = s.a[:0]
}

func (s *Slice) Get(i int) *roaring.Bitmap {
	if i < len(s.a) {
		return s.a[i]
	}
	return nil
}

func (s *Slice) GetOrCreate(i int) *roaring.Bitmap {
	if i < len(s.a) {
		return s.a[i]
	}
	b := NewBitmap()
	s.a = append(s.a, b)
	return b
}

// Export calls f with all bitmaps. idx 0 is for existence , followed by set bits.
func (b *BSI) Export(f func(idx int, bitmap *roaring.Bitmap)) {
	f(0, b.ex()) // existence bitmap goes first
	for i := 0; i < 64; i++ {
		e := b.get(i)
		if e == nil {
			break
		}
		f(i, e)
	}
}

type KV []*roaring.Bitmap

var _ Source = (*KV)(nil)

func (kv KV) BitCount() int {
	if len(kv) == 0 {
		return 0
	}
	return len(kv) - 1
}

func (kv KV) GetOrCreate(i int) *roaring.Bitmap { return nil }

func (kv KV) Get(i int) *roaring.Bitmap {
	if i < len(kv) {
		return kv[i]
	}
	return nil
}
