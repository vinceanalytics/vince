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

	for cu.Next() {
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
