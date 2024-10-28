package timeseries

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/dgryski/go-farm"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/util/assert"
	"github.com/vinceanalytics/vince/internal/util/oracle"
	"github.com/vinceanalytics/vince/internal/util/tree"
)

type translation struct {
	id       uint64
	tree     *tree.Tree
	ranges   [models.TranslatedFieldsSize]uint64
	keys     [models.TranslatedFieldsSize][][]byte
	values   [models.TranslatedFieldsSize][]uint64
	onAssign func(key []byte, uid uint64)
}

func (tr *translation) Release() error {
	return tr.tree.Close()
}

func newTranslation(db *pebble.DB, onAssign func(key []byte, uid uint64)) *translation {
	tr := translation{tree: tree.NewTree(oracle.DataPath), onAssign: onAssign}
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
			hash := farm.Fingerprint64(key)
			value := binary.BigEndian.Uint64(iter.Value())
			tr.tree.Set(hash, value)
			if onAssign != nil {
				onAssign(key, value)
			}
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
	hash := farm.Fingerprint64(key)
	uid := tr.tree.Get(hash)
	if uid > 0 {
		return uid
	}
	idx := field.Mutex()
	tr.ranges[idx]++
	uid = tr.ranges[idx]

	tr.tree.Set(hash, uid)
	tr.keys[idx] = append(tr.keys[idx], value)
	tr.values[idx] = append(tr.values[idx], uid)
	if tr.onAssign != nil {
		tr.onAssign(key, uid)
	}
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
