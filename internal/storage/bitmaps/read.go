package bitmaps

import (
	"github.com/gernest/roaring"
)

const (
	exponent                 = 20
	shardWidth               = 1 << exponent
	rowExponent              = (exponent - 16)  // for instance, 20-16 = 4
	RowWidth                 = 1 << rowExponent // containers per row, for instance 1<<4 = 16
	KeyMask                  = (RowWidth - 1)   // a mask for offset within the row
	ShardVsContainerExponent = 4
	RowMask                  = ^roaring.FilterKey(KeyMask) // a mask for the row bits, without converting them to a row ID
)

type Reader interface {
	Iter
	Max() uint64
	OffsetRange(offset, start, end uint64) *roaring.Bitmap
}

type Iter interface {
	Seek(key uint64) bool
	Valid() bool
	Next() bool
	Value() (uint64, *roaring.Container)
}

type OffsetRanger interface {
	OffsetRange(offset, start, end uint64) *roaring.Bitmap
}

type Wrap struct {
	roaring.ContainerIterator
	valid bool
}

var _ Iter = (*Wrap)(nil)

func (w *Wrap) Seek(_ uint64) bool {
	return w.Next()
}

func (w *Wrap) Valid() bool {
	return w.valid
}

func (w *Wrap) Next() bool {
	w.valid = w.ContainerIterator.Next()
	return w.valid
}

func Row(ra OffsetRanger, shard, rowID uint64) *roaring.Bitmap {
	return ra.OffsetRange(shardWidth*shard, shardWidth*rowID, shardWidth*(rowID+1))
}

func Existence(tx OffsetRanger, shard uint64) *roaring.Bitmap {
	return Row(tx, shard, bsiExistsBit)
}
