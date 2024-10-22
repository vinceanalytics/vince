package timeseries

import (
	"context"
	"sync"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/util/data"
	"github.com/vinceanalytics/vince/internal/util/lru"
	xt "github.com/vinceanalytics/vince/internal/util/translation"
	"github.com/vinceanalytics/vince/internal/util/trie"
)

type Timeseries struct {
	db *pebble.DB
	ba *batch

	mu   sync.RWMutex
	trie *trie.Trie

	// Bitmaps are organized that  similar queries depending on interval will always
	// fetch similar bitmaps.
	//
	// we need to have an up to date state of the cache to avoid missing new data.
	// so, cache.ra enumerate all cached bitmpas that we *And after each batch to
	// detect what keys to invalidate
	cache struct {
		mu  sync.RWMutex
		ra  *roaring.Bitmap
		lru *lru.Cache[uint64, *roaring.Bitmap]
	}
}

func New(db *pebble.DB) *Timeseries {
	ts := &Timeseries{db: db, trie: trie.NewTrie()}
	tr := newTranslation(db, ts.trie.Put)
	tr.onAssign = func(key []byte, uid uint64) {
		ts.mu.Lock()
		ts.trie.Put(key, uid)
		ts.mu.Unlock()
	}
	ts.ba = newbatch(db, tr)
	ts.cache.lru = lru.New[uint64, *roaring.Bitmap](8 << 10)
	ts.cache.ra = roaring.NewBitmap()
	return ts
}

var _ xt.Translator = (*Timeseries)(nil)

func (ts *Timeseries) Translate(field models.Field, value []byte) uint64 {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.trie.Get(encoding.TranslateKey(field, value))
}

func (ts *Timeseries) Get() *pebble.DB {
	return ts.db
}

func (ts *Timeseries) Close() error {
	return ts.Save()
}

func (ts *Timeseries) Save() error {
	defer ts.ba.keys.Reset()
	err := ts.ba.save()
	if err != nil {
		return err
	}
	ts.cache.mu.RLock()
	// inline intersection because we will discard  keys when we are done. No
	// need to acquire write lock. Most of the cases there will be no fresh cached
	// keys. Usage is mostly heavy writes and rare reads
	ts.ba.keys.And(ts.cache.ra)
	ts.cache.mu.RUnlock()
	if ts.ba.keys.IsEmpty() {
		return nil
	}
	ts.cache.mu.Lock()
	ts.ba.keys.Each(ts.cache.lru.Remove)
	ts.cache.mu.Unlock()
	return nil
}

func (ts *Timeseries) Add(m *models.Model) error {
	return ts.ba.add(m)
}

func (ts *Timeseries) NewBitmap(ctx context.Context, shard uint64, view uint64, field models.Field) (b *roaring.Bitmap) {
	buf := make([]byte, encoding.BitmapKeySize)
	key := encoding.Bitmap(shard, view, field, buf)
	keyHash := hash(key)

	// load from cache first
	ts.cache.mu.RLock()
	b, ok := ts.cache.lru.Get(keyHash)
	ts.cache.mu.RUnlock()
	if ok {
		return
	}
	data.Get(ts.db, key, func(val []byte) error {
		b = roaring.FromBufferWithCopy(val)
		ts.cache.mu.Lock()
		ts.cache.lru.Add(keyHash, b)
		ts.cache.ra.Set(keyHash)
		ts.cache.mu.Unlock()
		return nil
	})
	return
}
