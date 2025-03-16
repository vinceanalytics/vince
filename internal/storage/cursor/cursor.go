package cursor

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/cockroachdb/pebble"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/storage/bitmaps"
	"github.com/vinceanalytics/vince/internal/storage/fields"
)

type Cursor struct {
	it *pebble.Iterator
	lo fields.DataKey
	hi fields.DataKey
}

var _ bitmaps.Reader = (*Cursor)(nil)

func New(it *pebble.Iterator, shard uint64) *Cursor {
	return &Cursor{
		it: it,
	}
}

// ResetData seeks to the first data bitmap container for shard / field . Returns true
// if a matching container was found.
func (cu *Cursor) ResetData(field v1.Field, kind v1.DataType, shard uint64) bool {
	cu.lo.Make(field, kind, shard, 0)
	cu.lo.Make(field, kind, shard, math.MaxUint64)
	return cu.it.SeekGE(cu.lo[:]) && cu.Valid()
}

// First is a wrapper for Seek(0).
func (cu *Cursor) First() bool {
	return cu.Seek(0)
}

// Valid returns true if the cursor is at a valid container belonging to the same field
// we positioned with ResetData or ResetExistence.
func (cu *Cursor) Valid() bool {
	return cu.it.Valid() &&
		bytes.Compare(cu.it.Key(), cu.hi[:]) == -1
}

// Next advances the cursor to the next container. Return true if a valid container was found.
func (cu *Cursor) Next() bool {
	ok := cu.it.Next() && cu.Valid()
	if !ok {
		return false
	}
	// update container key
	copy(cu.lo[fields.ContainerOffset:], cu.it.Key()[fields.ContainerOffset:])
	return true
}

// Key returns container Key of the current position of the cursor.
func (cu *Cursor) Key() uint64 {
	key := cu.it.Key()
	return binary.BigEndian.Uint64(key[fields.ContainerOffset:])
}

// Value returns current container and its key pointed by the cursor.
func (cu *Cursor) Value() (uint64, *roaring.Container) {
	return cu.Key(), roaring.DecodeContainer(cu.it.Value())
}

// Container returns container  of the current position of the cursor.
func (cu *Cursor) Container() *roaring.Container {
	return roaring.DecodeContainer(cu.it.Value())
}

// Max returns maximum value recorded by the field bitmap. It is mainly used to detect
// bit depth with BSI encoded fields.
func (cu *Cursor) Max() uint64 {
	if !cu.it.SeekLT(cu.hi[:]) {
		return 0
	}
	key := cu.it.Key()
	if bytes.Compare(key, cu.lo[:]) == -1 {
		return 0
	}
	ck := binary.BigEndian.Uint64(key[fields.ContainerOffset:])
	value := roaring.LastValueFromEncodedContainer(cu.it.Value())
	return ((ck << 16) | uint64(value))
}

// Seek moves the iterator cursor to container matching key or the nearest container to key.
func (cu *Cursor) Seek(key uint64) bool {
	cu.lo.SetContainer(key)
	return cu.it.SeekGE(cu.lo[:]) && cu.Valid()
}

func (cu *Cursor) OffsetRange(offset, start, endx uint64) *roaring.Bitmap {
	other := roaring.NewSliceBitmap()
	off := highbits(offset)
	hi0, hi1 := highbits(start), highbits(endx)
	if !cu.Seek(hi0) {
		return other
	}
	for ; cu.Valid(); cu.it.Next() {
		key := cu.it.Key()
		ckey := binary.BigEndian.Uint64(key[fields.ContainerOffset:])
		if ckey >= hi1 {
			break
		}
		other.Containers.Put(off+(ckey-hi0), roaring.DecodeContainer(cu.it.Value()).Clone())
	}
	return other
}

func (cu *Cursor) ReadMutex(field v1.Field, shard uint64, match *roaring.Bitmap) [][]uint64 {
	if !cu.ResetData(field, v1.DataType_mutex, shard) {
		return make([][]uint64, match.Count())
	}

	ex := bitmaps.NewExtractor()
	defer ex.Release()
	return ex.Mutex(cu, shard, match)
}

func (cu *Cursor) ReadBSI(field v1.Field, shard uint64, match *roaring.Bitmap) []uint64 {
	if !cu.ResetData(field, v1.DataType_bsi, shard) {
		return make([]uint64, match.Count())
	}

	ex := bitmaps.NewExtractor()
	defer ex.Release()

	depth := cu.Max() / shardwidth.ShardWidth
	return ex.BSI(cu, depth, shard, match)
}

func highbits(v uint64) uint64 { return v >> 16 }
