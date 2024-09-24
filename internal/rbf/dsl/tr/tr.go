package tr

import (
	"bytes"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/blevesearch/vellum"
	"github.com/blevesearch/vellum/regexp"
	"go.etcd.io/bbolt"
)

var (
	keys     = []byte("keys")
	ids      = []byte("ids")
	blobHash = []byte("blob_hash")
	blobID   = []byte("blob_id")
	fst      = []byte("fst")
)

var emptyKey = []byte{
	0x00, 0x00, 0x00,
	0x4d, 0x54, 0x4d, 0x54, // MTMT
	0x00,
	0xc2, 0xa0, // NO-BREAK SPACE
	0x00,
}

type File struct {
	db   *bbolt.DB
	path string
	skip map[string]struct{}
}

func New(path string, skip ...string) *File {
	m := map[string]struct{}{}
	for _, n := range skip {
		m[n] = struct{}{}
	}
	return &File{path: path, skip: m}
}

func (f *File) Open() error {
	db, err := bbolt.Open(f.path, 0600, nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(keys)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(ids)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(blobHash)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(blobID)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(fst)
		return err
	})
	if err != nil {
		db.Close()
		return err
	}
	f.db = db
	return nil
}

func (f *File) Close() (err error) {
	if f.db != nil {
		err = f.db.Close()
		f.db = nil
	}
	return
}

func (f *File) Write() (*Write, error) {
	tx, err := f.db.Begin(true)
	if err != nil {
		return nil, err
	}
	return &Write{
		tx:       tx,
		keys:     tx.Bucket(keys),
		ids:      tx.Bucket(ids),
		blobID:   tx.Bucket(blobID),
		blobHash: tx.Bucket(blobHash),
		fst:      tx.Bucket(fst),
	}, nil
}

type Write struct {
	tx       *bbolt.Tx
	keys     *bbolt.Bucket
	ids      *bbolt.Bucket
	blobID   *bbolt.Bucket
	blobHash *bbolt.Bucket
	fst      *bbolt.Bucket
}

func (w *Write) Release() error {
	return w.tx.Rollback()
}

func (w *Write) Commit() error {
	err := w.vellum()
	if err != nil {
		return err
	}
	return w.tx.Commit()
}

var builderPool = &sync.Pool{New: func() any {
	b, _ := vellum.New(io.Discard, nil)
	return b
}}

func (w *Write) vellum() error {
	var o bytes.Buffer
	b := builderPool.Get().(*vellum.Builder)
	defer func() {
		b.Reset(io.Discard)
		builderPool.Put(b)
	}()
	return w.keys.ForEachBucket(func(k []byte) error {
		o.Reset()
		err := b.Reset(&o)
		if err != nil {
			return err
		}
		err = w.keys.Bucket(k).ForEach(func(k, v []byte) error {
			if bytes.Equal(k, emptyKey) {
				// skip empty keys
				return nil
			}
			return b.Insert(k, binary.BigEndian.Uint64(v))
		})
		if err != nil {
			return err
		}
		err = b.Close()
		if err != nil {
			return err
		}
		return w.fst.Put(k, bytes.Clone(o.Bytes()))
	})
}

func (w *Write) String(field string) (*String, error) {
	keys, err := bucket(w.keys, []byte(field))
	if err != nil {
		return nil, fmt.Errorf("ebf/tr: setup keys bucket %w", err)
	}
	ids, err := bucket(w.ids, []byte(field))
	if err != nil {
		return nil, fmt.Errorf("ebf/tr: setup ids bucket %w", err)
	}
	return &String{keys: keys, ids: ids}, nil
}

func (w *Write) Blobs(field string) (*Blob, error) {
	keys, err := bucket(w.blobHash, []byte(field))
	if err != nil {
		return nil, fmt.Errorf("ebf/tr: setup keys bucket %w", err)
	}
	ids, err := bucket(w.blobID, []byte(field))
	if err != nil {
		return nil, fmt.Errorf("ebf/tr: setup ids bucket %w", err)
	}
	return &Blob{keys: keys, ids: ids}, nil
}

type String struct {
	keys *bbolt.Bucket
	ids  *bbolt.Bucket
	b    [8]byte
}

func (c *String) Bulk(keys []string, result []uint64) error {
	for i := range keys {
		key := []byte(keys[i])
		if len(key) == 0 {
			key = emptyKey
		}
		// fast path: hey already translated.
		if value := c.keys.Get(key); value != nil {
			result[i] = binary.BigEndian.Uint64(value)
			continue
		}
		next, err := c.ids.NextSequence()
		if err != nil {
			return fmt.Errorf("ebf/tr: getting seq id %w", err)
		}

		binary.BigEndian.PutUint64(c.b[:], next)

		err = c.keys.Put(key, bytes.Clone(c.b[:]))
		if err != nil {
			return fmt.Errorf("ebf/tr: writing key %w", err)
		}
		err = c.ids.Put(c.b[:], key)
		if err != nil {
			return fmt.Errorf("ebf/tr: writing key %w", err)
		}
		result[i] = next
	}
	return nil
}

func (c *String) BulkSet(keys [][]string, result [][]uint64) error {
	for n := range keys {
		for i := range keys[n] {
			key := []byte(keys[n][i])
			if len(key) == 0 {
				key = emptyKey
			}
			// fast path: hey already translated.
			if value := c.keys.Get(key); value != nil {
				result[n][i] = binary.BigEndian.Uint64(value)
				continue
			}
			next, err := c.ids.NextSequence()
			if err != nil {
				return fmt.Errorf("ebf/tr: getting seq id %w", err)
			}

			binary.BigEndian.PutUint64(c.b[:], next)

			err = c.keys.Put(key, bytes.Clone(c.b[:]))
			if err != nil {
				return fmt.Errorf("ebf/tr: writing key %w", err)
			}
			err = c.ids.Put(c.b[:], key)
			if err != nil {
				return fmt.Errorf("ebf/tr: writing key %w", err)
			}
			result[n][i] = next
		}
	}
	return nil
}

type Blob struct {
	keys *bbolt.Bucket
	ids  *bbolt.Bucket
	b    [8]byte
}

func (c *Blob) Bulk(keys [][]byte, result []uint64) error {
	for i := range keys {
		key := keys[i]
		if len(key) == 0 {
			key = emptyKey
		}
		hash := sha512.Sum512_224(key)
		if value := c.keys.Get(hash[:]); value != nil {
			result[i] = binary.BigEndian.Uint64(value)
			continue
		}
		next, err := c.ids.NextSequence()
		if err != nil {
			return fmt.Errorf("ebf/tr: getting seq id %w", err)
		}
		binary.BigEndian.PutUint64(c.b[:], next)
		err = c.keys.Put(hash[:], bytes.Clone(c.b[:]))
		if err != nil {
			return fmt.Errorf("ebf/tr: writing blob key %w", err)
		}
		err = c.ids.Put(c.b[:], key)
		if err != nil {
			return fmt.Errorf("ebf/tr: writing blob id %w", err)
		}
		result[i] = next
	}
	return nil
}

func (c *Blob) BulkSet(keys [][][]byte, result [][]uint64) error {
	for n := range keys {
		for i := range keys[n] {
			key := keys[n][i]
			if len(key) == 0 {
				key = emptyKey
			}
			hash := sha512.Sum512_224(key)
			if value := c.keys.Get(hash[:]); value != nil {
				result[n][i] = binary.BigEndian.Uint64(value)
				continue
			}
			next, err := c.ids.NextSequence()
			if err != nil {
				return fmt.Errorf("ebf/tr: getting seq id %w", err)
			}
			binary.BigEndian.PutUint64(c.b[:], next)
			err = c.keys.Put(hash[:], bytes.Clone(c.b[:]))
			if err != nil {
				return fmt.Errorf("ebf/tr: writing blob key %w", err)
			}
			err = c.ids.Put(c.b[:], key)
			if err != nil {
				return fmt.Errorf("ebf/tr: writing blob id %w", err)
			}
			result[n][i] = next
		}
	}
	return nil
}

func (f *File) Read() (*Read, error) {
	tx, err := f.db.Begin(false)
	if err != nil {
		return nil, err
	}
	return &Read{
		tx:       tx,
		keys:     tx.Bucket(keys),
		ids:      tx.Bucket(ids),
		fst:      tx.Bucket(fst),
		blobID:   tx.Bucket(blobID),
		blobHash: tx.Bucket(blobHash),
	}, nil
}

type Read struct {
	tx       *bbolt.Tx
	keys     *bbolt.Bucket
	ids      *bbolt.Bucket
	fst      *bbolt.Bucket
	blobID   *bbolt.Bucket
	blobHash *bbolt.Bucket
}

func (r *Read) Release() error {
	return r.tx.Rollback()
}

func (r *Read) Key(field string, id uint64) []byte {
	ids := r.ids.Bucket([]byte(field))
	if ids == nil {
		return nil
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], id)
	return get(ids.Get(b[:]))
}

func (r *Read) Keys(field string, id []uint64, f func(value []byte)) {
	ids := r.ids.Bucket([]byte(field))
	if ids == nil {
		return
	}
	var b [8]byte
	for i := range id {
		binary.BigEndian.PutUint64(b[:], id[i])
		f(get(ids.Get(b[:])))
	}
}

func get(key []byte) []byte {
	if bytes.Equal(key, emptyKey) {
		return []byte{}
	}
	return key
}

func (r *Read) Find(field string, key []byte) (uint64, bool) {
	if len(key) == 0 {
		key = emptyKey
	}
	keys := r.keys.Bucket([]byte(field))
	if keys == nil {
		return 0, false
	}
	value := keys.Get(key)
	if value != nil {
		return binary.BigEndian.Uint64(value), true
	}
	return 0, false
}

// Blob returns data stored for blob id.
func (r *Read) Blob(field string, id uint64) []byte {
	ids := r.blobID.Bucket([]byte(field))
	if ids == nil {
		return nil
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], id)
	return get(ids.Get(b[:]))
}

func (r *Read) FindBlob(field string, blob []byte) (uint64, bool) {
	if len(blob) == 0 {
		blob = emptyKey
	}
	b := r.blobHash.Bucket([]byte(field))
	if b == nil {
		return 0, false
	}
	hash := sha512.Sum512_224(blob)
	value := b.Get(hash[:])
	if value != nil {
		return binary.BigEndian.Uint64(value), true
	}
	return 0, false
}

func (r *Read) SearchRe(field string, like string, start, end []byte, match func(key []byte, value uint64) error) error {
	b := r.fst.Get([]byte(field))
	if b == nil {
		return nil
	}
	re, err := regexp.New(like)
	if err != nil {
		return fmt.Errorf("compiling regex %w", err)
	}
	fst, err := vellum.Load(b)
	if err != nil {
		return err
	}
	it, err := fst.Search(re, start, end)
	for err == nil {
		err = match(it.Current())
		if err != nil {
			return err
		}
		err = it.Next()
	}
	return nil
}
func (r *Read) Search(field string, a vellum.Automaton, start, end []byte, match func(key []byte, value uint64) error) error {
	b := r.fst.Get([]byte(field))
	if b == nil {
		return nil
	}
	fst, err := vellum.Load(b)
	if err != nil {
		return err
	}
	it, err := fst.Search(a, start, end)
	for err == nil {
		err = match(it.Current())
		if err != nil {
			return err
		}
		err = it.Next()
	}
	return nil
}

func bucket(b *bbolt.Bucket, key []byte) (*bbolt.Bucket, error) {
	if v := b.Bucket(key); v != nil {
		return v, nil
	}
	return b.CreateBucket(key)
}
