package timeseries

import "github.com/RoaringBitmap/roaring/roaring64"

type Uniq struct {
	hash *roaring64.Bitmap
}

func NewUniq() *Uniq {
	return &Uniq{
		hash: roaring64.New(),
	}
}

func (u *Uniq) Has(uid uint64) bool {
	if u.hash.Contains(uid) {
		return true
	}
	u.hash.Add(uid)
	return false
}
