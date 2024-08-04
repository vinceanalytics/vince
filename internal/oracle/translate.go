package oracle

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"regexp"

	"go.etcd.io/bbolt"
)

var (
	keys     = []byte("k")
	ids      = []byte("i")
	emptyKey = []byte{
		0x00, 0x00, 0x00,
		0x4d, 0x54, 0x4d, 0x54, // MTMT
		0x00,
		0xc2, 0xa0, // NO-BREAK SPACE
		0x00,
	}
)

type field struct {
	bucket *bbolt.Bucket
	keys   *bbolt.Bucket
	ids    *bbolt.Bucket
}

func newWriteField(tx *bbolt.Tx, name []byte) (*field, error) {
	b, err := tx.CreateBucketIfNotExists(name)
	if err != nil {
		return nil, err
	}
	kb, err := b.CreateBucketIfNotExists(keys)
	if err != nil {
		return nil, err
	}
	ib, err := b.CreateBucketIfNotExists(ids)
	if err != nil {
		return nil, err
	}
	return &field{bucket: b, keys: kb, ids: ib}, nil
}

func (f *field) translate(value []byte) (uint64, error) {
	if len(value) == 0 {
		value = emptyKey
	}
	if v := f.keys.Get(value); v != nil {
		return binary.BigEndian.Uint64(v), nil
	}
	seq, err := f.bucket.NextSequence()
	if err != nil {
		return 0, fmt.Errorf("next sequence%w", err)
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], seq)
	err = f.keys.Put(value, b[:])
	if err != nil {
		return 0, err
	}
	err = f.ids.Put(b[:], value)
	if err != nil {
		return 0, err
	}
	return seq, nil
}

func newReadField(tx *bbolt.Tx, name []byte) *field {
	if b := tx.Bucket(name); b != nil {
		return &field{bucket: b, ids: b.Bucket(ids), keys: b.Bucket(keys)}
	}
	return &field{}
}

func (f *field) read(id uint64) []byte {
	if f.bucket != nil {
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], id)
		if v := f.ids.Get(b[:]); v != nil {
			if bytes.Equal(emptyKey, v) {
				return []byte{}
			}
			return v
		}
	}
	return []byte{}
}

func (f *field) get(value []byte) (id uint64, ok bool) {
	if f.bucket != nil {
		if v := f.keys.Get(value); v != nil {
			return binary.BigEndian.Uint64(v), true
		}
	}
	return
}

func (f *field) search(re *regexp.Regexp) (o []uint64) {
	if f.bucket != nil {
		f.keys.ForEach(func(k, v []byte) error {
			if re.Match(k) {
				o = append(o, binary.BigEndian.Uint64(v))
			}
			return nil
		})
	}
	return
}
