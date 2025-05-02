package cursor

import (
	"bytes"

	"github.com/cockroachdb/pebble"
	"github.com/gernest/roaring"
)

type Batcher interface {
	Set(key, value []byte, _ *pebble.WriteOptions) error
	Delete(key []byte, _ *pebble.WriteOptions) error
}

func (cu *Cursor) ApplyFilter(key uint64, filter roaring.BitmapFilter) error {
	var minKey roaring.FilterKey
	cu.Seek(key)
	for cu.Valid() {
		key := roaring.FilterKey(cu.Key())
		if key < minKey {
			cu.Next()
			continue
		}
		co := cu.Container()
		res := filter.ConsiderKey(key, co.N())
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
			cu.Seek(uint64(minKey))
		} else {
			cu.Next()
		}
	}
	return nil
}

func (cu *Cursor) ClearRecords(ba Batcher, columns *roaring.Bitmap) error {
	rewriteExisting := roaring.NewBitmapBitmapTrimmer(columns, func(key roaring.FilterKey, data, filter *roaring.Container, writeback roaring.ContainerWriteback) error {
		if filter.N() == 0 {
			return nil
		}
		existing := data.N()
		// nothing to delete. this can't happen normally, but the rewriter calls
		// us with an empty data container when it's done.
		if existing == 0 {
			return nil
		}
		diff := data.Difference(filter)
		if changed(diff, data) {
			return writeback(key, diff)
		}
		return nil
	})
	return cu.ApplyRewriter(ba, rewriteExisting)
}

func changed(diff, orig *roaring.Container) bool {
	if diff.N() == 0 {
		return true
	}
	if diff.N() != orig.N() {
		return true
	}
	a := diff.Slice()
	b := orig.Slice()
	for i := range a {
		if a[i] != b[i] {
			return true
		}
	}
	return false
}

func (cu *Cursor) ApplyRewriter(ba Batcher, rewriter roaring.BitmapRewriter) error {
	var (
		minKey roaring.FilterKey
		dirty  bool
	)
	if rewriter == nil {
		return nil
	}
	var writeback roaring.ContainerWriteback = func(updateKey roaring.FilterKey, data *roaring.Container) (err error) {
		dirty = true
		var exact bool
		if cu.Key() == uint64(updateKey) {
			dirty = false
			exact = true
		} else {
			exact = cu.seek(uint64(updateKey))
		}
		if data.N() == 0 {
			if exact {
				err = ba.Delete(cu.lo[:], nil)
			}
			// if we don't delete, we aren't changing our situation at all
		} else {
			err = ba.Set(cu.lo[:], data.Encode(), nil)
		}
		return err
	}
	cu.First()
	for cu.Valid() {
		key := roaring.FilterKey(cu.Key())
		if key < minKey {
			continue
		}
		co := cu.Container()
		res := rewriter.ConsiderKey(key, co.N())
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
		} else {
			cu.Next()
		}
	}
	res := rewriter.RewriteData(^roaring.FilterKey(0), nil, writeback)
	return res.Err
}

func (cu *Cursor) seek(key uint64) bool {
	return cu.Seek(key) &&
		bytes.Equal(cu.it.Key(), cu.lo[:])
}
