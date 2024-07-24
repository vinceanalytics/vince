package len64

import "github.com/RoaringBitmap/roaring/v2/roaring64"

func Visitors(db *View, foundSet *roaring64.Bitmap) (uint64, error) {
	uid, err := db.Get("uid")
	if err != nil {
		return 0, err
	}
	uniq := uid.TransposeWithCounts(
		parallel(), foundSet, foundSet)
	return uniq.GetCardinality(), nil
}

func Visits(db *View, foundSet *roaring64.Bitmap) (uint64, error) {
	session, err := db.Get("session")
	if err != nil {
		return 0, err
	}
	sum, _ := session.Sum(foundSet)
	return uint64(sum), nil
}

func Bounce(db *View, foundSet *roaring64.Bitmap) (uint64, error) {
	bounce, err := db.Get("bounce")
	if err != nil {
		return 0, err
	}
	sum, _ := bounce.Sum(foundSet)
	return uint64(sum), nil
}
