package timeseries

import (
	"context"
	"iter"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/keys"
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

	trie struct {
		mu sync.RWMutex
		tr *trie.Trie
	}

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

	// To avoid blindly iterating on all shards to find views. We keep a mapping
	// of observed views in a shard.
	//
	// We use a slice because shards starts from 0 and there is no gaps.
	views struct {
		mu sync.RWMutex
		ra []*roaring.Bitmap
	}
}

func New(db *pebble.DB) *Timeseries {
	ts := &Timeseries{db: db}
	ts.trie.tr = trie.NewTrie()
	tr := newTranslation(db, ts.trie.tr.Put)
	tr.onAssign = func(key []byte, uid uint64) {
		ts.trie.mu.Lock()
		ts.trie.tr.Put(key, uid)
		ts.trie.mu.Unlock()
	}
	ts.ba = newbatch(db, tr)
	ts.cache.lru = lru.New[uint64, *roaring.Bitmap](8 << 10)
	ts.cache.ra = roaring.NewBitmap()

	// load all views into memory. We append to  ts.views.ra because shards are always
	// sorted and starts with 0.
	data.Prefix(db, keys.ShardsPrefix, func(key, value []byte) error {
		ra := roaring.FromBufferWithCopy(value)
		ts.views.ra = append(ts.views.ra, ra)
		return nil
	})
	return ts
}

var _ xt.Translator = (*Timeseries)(nil)

func (ts *Timeseries) Translate(field models.Field, value []byte) uint64 {
	ts.trie.mu.RLock()
	defer ts.trie.mu.RUnlock()
	return ts.trie.tr.Get(encoding.TranslateKey(field, value))
}

func (ts *Timeseries) Get() *pebble.DB {
	return ts.db
}

func (ts *Timeseries) Close() error {
	return ts.Save()
}

func (ts *Timeseries) Save() error {
	err := ts.ba.save()
	if err != nil {
		return err
	}
	ts.updateCache()
	ts.updateViews()
	return nil
}

func (ts *Timeseries) Add(m *models.Model) error {
	return ts.ba.add(m)
}

func (ts *Timeseries) NewBitmap(ctx context.Context, shard uint64, view uint64, field models.Field) (b *roaring.Bitmap) {
	key := encoding.Bitmap(shard, view, field)
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

func (ts *Timeseries) Shards(views iter.Seq[time.Time]) iter.Seq2[uint64, uint64] {
	ra := roaring.NewBitmap()
	for v := range views {
		ra.Set(uint64(v.UnixMilli()))
	}
	rs := make([]*roaring.Bitmap, len(ts.views.ra))
	ts.views.mu.RLock()
	for i := range rs {
		clone := ra.Clone()
		clone.And(ts.views.ra[i])
		rs[i] = clone
	}
	ts.views.mu.RUnlock()

	return func(yield func(uint64, uint64) bool) {
		for i := range rs {
			ok := rs[i].EachOk(func(value uint64) bool {
				return yield(uint64(i), value)
			})
			if !ok {
				return
			}
		}
	}
}

func (ts *Timeseries) Global() iter.Seq2[uint64, uint64] {
	ts.views.mu.RLock()
	shards := len(ts.views.ra)
	ts.views.mu.RUnlock()
	return func(yield func(uint64, uint64) bool) {
		for shard := range shards {
			if !yield(uint64(shard), 0) {
				return
			}
		}
	}
}

func (ts *Timeseries) updateCache() {
	defer ts.ba.keys.Reset()
	ts.cache.mu.RLock()
	// inline intersection because we will discard  keys when we are done. No
	// need to acquire write lock. Most of the cases there will be no fresh cached
	// keys. Usage is mostly heavy writes and rare reads
	ts.ba.keys.And(ts.cache.ra)
	ts.cache.mu.RUnlock()
	if ts.ba.keys.IsEmpty() {
		return
	}
	ts.cache.mu.Lock()
	ts.ba.keys.Each(ts.cache.lru.Remove)
	ts.cache.mu.Unlock()
}

func (ts *Timeseries) updateViews() {
	ts.views.mu.Lock()
	defer ts.views.mu.Unlock()
	if ts.ba.shard < uint64(len(ts.views.ra)) {
		ts.views.ra[ts.ba.shard].Or(ts.ba.views)
	} else {
		ts.views.ra = append(ts.views.ra, ts.ba.views.Clone())
	}
}
