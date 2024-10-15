package bsi

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
)

type KV struct {
	key    []byte
	mu     sync.RWMutex
	tx     *badger.Txn
	cached map[int]*roaring.Bitmap
}

func NewKV(tx *badger.Txn, ts, shard uint64, field models.Field) *BSI {
	key := encoding.Bitmap(ts, shard, field, 0,
		make([]byte, encoding.BitmapKeySize))
	kv := &KV{
		key:    key,
		tx:     tx,
		cached: make(map[int]*roaring.Bitmap),
	}
	return &BSI{source: kv}
}

var _ Source = (*KV)(nil)

func (kv *KV) GetOrCreate(i int) *roaring.Bitmap { return nil }

func (kv *KV) Get(i int) *roaring.Bitmap {
	kv.mu.RLock()
	b, ok := kv.cached[i]
	kv.mu.RUnlock()
	if ok {
		return b
	}
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.key[len(kv.key)-1] = byte(i)
	it, err := kv.tx.Get(kv.key)
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			slog.Error("reading bitmap", "err", err)
		}
		return nil
	}
	value, err := it.ValueCopy(nil)
	if err != nil {
		slog.Error("copying bitmap", "err", err)
		return nil
	}
	b = roaring.FromBuffer(value)
	kv.cached[i] = b
	return b
}
