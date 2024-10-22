package roaring

import "io"

func (ra *Bitmap) MergeNewer(value []byte) error {
	ra.Or(FromBuffer(value))
	return nil
}

func (ra *Bitmap) MergeOlder(value []byte) error {
	ra.Or(FromBuffer(value))
	return nil
}

func (ra *Bitmap) Finish(includesBase bool) ([]byte, io.Closer, error) {
	return ra.ToBuffer(), nil, nil
}
