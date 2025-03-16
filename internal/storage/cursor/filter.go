package cursor

import (
	"bytes"

	"github.com/cockroachdb/pebble"
	"github.com/gernest/roaring"
)

type Batcher interface {
	Merge(key []byte, value []byte, _ *pebble.WriteOptions) error
	Delete(key []byte, _ *pebble.WriteOptions) error
}

func (cu *Cursor) ApplyFilter(key uint64, filter roaring.BitmapFilter) error {
	var minKey roaring.FilterKey
	for cu.Seek(key); cu.Valid(); cu.Next() {
		key := roaring.FilterKey(cu.Key())
		if key < minKey {
			continue
		}
		co := cu.Container()
		res := filter.ConsiderKey(key, int32(co.N()))
		if res.Err != nil {
			return res.Err
		}
		if res.YesKey <= key && res.NoKey <= key {
			res = filter.ConsiderData(key, co)
			if res.Err != nil {
				return res.Err
			}
		}
		minKey = res.NoKey
		if minKey > key+1 {
			if !cu.Seek(uint64(minKey)) {
				break
			}
		}
	}
	return nil
}

func (cu *Cursor) ClearRecords(ba Batcher, columns *roaring.Bitmap) error {
	rewriteExisting := roaring.NewBitmapBitmapTrimmer(columns, func(key roaring.FilterKey, data *roaring.Container, filter *roaring.Container, writeback roaring.ContainerWriteback) error {
		if filter.N() == 0 {
			return nil
		}
		existing := data.N()
		// nothing to delete. this can't happen normally, but the rewriter calls
		// us with an empty data container when it's done.
		if existing == 0 {
			return nil
		}
		data = data.DifferenceInPlace(filter)
		if data.N() != existing {
			return writeback(key, data)
		}
		return nil
	})
	return cu.ApplyRewriter(ba, rewriteExisting)
}

func (cu *Cursor) ApplyRewriter(ba Batcher, rewriter roaring.BitmapRewriter) error {
	var (
		minKey roaring.FilterKey
		dirty  bool
		key    roaring.FilterKey
	)
	if rewriter == nil {
		return nil
	}
	var writeback roaring.ContainerWriteback = func(updateKey roaring.FilterKey, data *roaring.Container) (err error) {
		dirty = true
		var exact bool
		if updateKey != key {
			exact = cu.seek(uint64(updateKey))
		} else {
			exact = true
		}
		if data.N() == 0 {
			if exact {
				err = ba.Delete(cu.lo[:], nil)
				key = ^roaring.FilterKey(0)
			}
			// if we don't delete, we aren't changing our situation at all
		} else {
			err = ba.Merge(cu.lo[:], data.Encode(), nil)
			key = ^roaring.FilterKey(0)
		}
		return err
	}

	for cu.First(); cu.Valid(); cu.Next() {
		key := roaring.FilterKey(cu.Key())
		if key < minKey {
			continue
		}
		co := cu.Container()
		res := rewriter.ConsiderKey(key, int32(co.N()))
		if res.Err != nil {
			return res.Err
		}
		if res.YesKey <= key && res.NoKey <= key {
			res = rewriter.RewriteData(key, co, writeback)
			if res.Err != nil {
				return res.Err
			}
		}
		// if the callback did any writing, we need to reset our cursor,
		// and if the next key is far away, we should also reset our cursor.
		//
		// In practice the "key+64" probably comes out to "we've been told
		// we're done".
		if dirty || minKey > (key+64) {
			dirty = false
			if cu.Seek(uint64(minKey)) {
				break
			}
		}
	}
	res := rewriter.RewriteData(^roaring.FilterKey(0), nil, writeback)
	return res.Err
}

func (cu *Cursor) seek(key uint64) bool {
	return cu.Seek(key) &&
		bytes.Equal(cu.it.Key(), cu.lo[:])
}
