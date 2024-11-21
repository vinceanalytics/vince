package data

import (
	"bytes"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/vinceanalytics/vince/internal/ro2"
)

func Open(path string, o *pebble.Options) (*pebble.DB, error) {
	if o == nil {
		o = &pebble.Options{}
	}
	if o.FS == nil {
		o.FS = vfs.Default
	}
	o.FormatMajorVersion = pebble.FormatColumnarBlocks
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

func PrefixKeys(db *pebble.DB, prefix []byte, f func(key []byte) error) error {
	iter, err := db.NewIter(nil)
	if err != nil {
		return err
	}
	defer iter.Close()
	for iter.SeekGE(prefix); iter.Valid(); iter.Next() {
		if !bytes.HasPrefix(iter.Key(), prefix) {
			break
		}
		err := f(iter.Key())
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

func Range(it *pebble.Iterator, lo, hi []byte, f func(key []byte, ra *ro2.Bitmap) error) error {
	ra := ro2.NewBitmap()
	for it.SeekGE(lo); it.Valid(); it.Next() {
		key := it.Key()
		if bytes.Compare(key, hi) != -1 {
			break
		}
		ra.Containers.Reset()
		err := ra.UnmarshalBinary(it.Value())
		if err != nil {
			return err
		}
		err = f(key, ra)
		if err != nil {
			return err
		}
	}
	return nil
}
