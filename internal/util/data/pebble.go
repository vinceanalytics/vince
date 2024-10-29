package data

import (
	"bytes"
	"path/filepath"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/vinceanalytics/vince/internal/roaring"
)

func Open(path string, o *pebble.Options) (*pebble.DB, error) {
	if path != "" {
		path = filepath.Join(path, "pebble")
	}
	if o == nil {
		o = new(pebble.Options)
		o.FS = vfs.Default
	}
	if path == "" {
		o.FS = vfs.NewMem()
	}
	o.Merger = BitmapMarger
	return pebble.Open(path, o)
}

func Prefix(db *pebble.DB, prefix []byte, f func(key, value []byte) error) error {
	iter, err := db.NewIter(nil)
	if err != nil {
		return err
	}
	defer iter.Close()
	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		if !bytes.HasPrefix(iter.Key(), prefix) {
			break
		}
		err := f(iter.Key(), iter.Value())
		if err != nil {
			return err
		}
	}
	return nil
}

func Has(db *pebble.DB, key []byte) bool {
	return Get(db, key, func(val []byte) error { return nil }) == nil
}

func Get(db *pebble.DB, key []byte, value func(val []byte) error) error {
	val, done, err := db.Get(key)
	if err != nil {
		return err
	}
	defer done.Close()
	return value(val)
}

func Range(it *pebble.Iterator, lo, hi []byte, f func(key []byte, ra *roaring.Bitmap) bool) {
	for it.SeekGE(lo); it.Valid(); it.Next() {
		key := it.Key()
		if bytes.Compare(key, hi) != -1 {
			break
		}
		if !f(key, roaring.FromBuffer(it.Value())) {
			break
		}
	}
}
