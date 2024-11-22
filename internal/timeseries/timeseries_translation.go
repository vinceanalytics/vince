package timeseries

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/swiss"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/util/assert"
	"github.com/vinceanalytics/vince/internal/util/hash"
)

type treeLocked struct {
	mu   sync.RWMutex
	tree *swiss.Map[uint64, uint64]
}

func newTree() *treeLocked {
	return &treeLocked{
		tree: swiss.New[uint64, uint64](1 << 19),
	}
}

func (t *treeLocked) Set(key []byte, value uint64) {
	hash := hash.Bytes(key)
	t.mu.Lock()
	t.tree.Put(hash, value)
	t.mu.Unlock()
}

func (t *treeLocked) Get(key []byte) (value uint64) {
	hash := hash.Bytes(key)
	t.mu.RLock()
	value, _ = t.tree.Get(hash)
	t.mu.RUnlock()
	return
}

func (c *treeLocked) Release() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tree.Clear()
	return nil
}

type translation struct {
	tree   *treeLocked
	keys   [models.TranslatedFieldsSize][][]byte
	values [models.TranslatedFieldsSize][]uint64
	ranges [models.TranslatedFieldsSize]uint64
	id     uint64
}

func newTranslation(db *pebble.DB, tree *treeLocked) *translation {
	tr := translation{tree: tree}
	iter, err := db.NewIter(&pebble.IterOptions{})
	assert.Nil(err, "openin iterator for translations")
	defer iter.Close()
	{
		// load  sequences
		for iter.SeekGE(keys.TranslateSeqPrefix); iter.Valid(); iter.Next() {
			key := iter.Key()
			if !bytes.HasPrefix(key, keys.TranslateSeqPrefix) {
				break
			}
			val := binary.BigEndian.Uint64(iter.Value())

			if len(key) == 2 {
				// global sequence records id
				tr.id = val
				continue
			}
			f := models.Field(key[2])
			tr.ranges[f.Mutex()] = val
		}
		start := time.Now()
		slog.Info("loading translation data")
		var count uint64
		// load translation
		for iter.SeekGE(keys.TranslateKeyPrefix); iter.Valid(); iter.Next() {
			key := iter.Key()
			if !bytes.HasPrefix(key, keys.TranslateKeyPrefix) {
				break
			}
			count++
			value := binary.BigEndian.Uint64(iter.Value())

			//  no need for locks for faster initialization
			tree.tree.Put(hash.Bytes(key), value)
		}
		slog.Info("complete loading translation",
			"elapsed", time.Since(start), "keys", count)

	}
	return &tr
}

func (tr *translation) Next() uint64 {
	tr.id++
	return tr.id
}

func (tr *translation) Assign(field models.Field, value []byte) uint64 {
	key := encoding.TranslateKey(field, value)
	uid := tr.tree.Get(key)
	if uid > 0 {
		return uid
	}
	idx := field.Mutex()
	tr.ranges[idx]++
	uid = tr.ranges[idx]

	tr.tree.Set(key, uid)
	tr.keys[idx] = append(tr.keys[idx], value)
	tr.values[idx] = append(tr.values[idx], uid)
	return uid
}

func (tr *translation) reset() {
	for i := range tr.keys {
		clear(tr.keys[i])
		tr.keys[i] = tr.keys[i][:0]
		clear(tr.values[i])
		tr.values[i] = tr.values[i][:0]
	}
}

func (tr *translation) flush(f func(key, value []byte, _ *pebble.WriteOptions) error) error {
	// write sequences first
	var b [8]byte
	key := make([]byte, 4<<10)
	key = key[:3]
	copy(key, keys.TranslateSeqPrefix)
	binary.BigEndian.PutUint64(b[:], tr.id)
	err := f(keys.TranslateSeqPrefix, b[:], nil)
	if err != nil {
		return fmt.Errorf("writing translation sequence key %w", err)
	}

	for i := range tr.ranges {
		key[2] = byte(models.Mutex(i))
		binary.BigEndian.PutUint64(b[:], tr.ranges[i])
		err := f(key, b[:], nil)
		if err != nil {
			return fmt.Errorf("writing translation sequence key %w", err)
		}
	}
	copy(key, keys.TranslateKeyPrefix)
	value := make([]byte, 3+8)
	copy(value, keys.TranslateIDPrefix)
	for i := range tr.keys {
		key[2] = byte(models.Mutex(i))
		value[2] = byte(models.Mutex(i))
		for j := range tr.keys[i] {
			key = append(key[:3], tr.keys[i][j]...)
			binary.BigEndian.PutUint64(value[3:], tr.values[i][j])
			err := f(key, value[3:], nil)
			if err != nil {
				return fmt.Errorf("writing translation key key %w", err)
			}
			err = f(value, key[3:], nil)
			if err != nil {
				return fmt.Errorf("writing translation id key %w", err)
			}
		}
	}
	return nil
}
