package timeseries

import (
	"iter"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/util/data"
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
	ts.updateViews()
	return nil
}

func (ts *Timeseries) Add(m *models.Model) error {
	return ts.ba.add(m)
}

func (ts *Timeseries) Shards(views iter.Seq[time.Time]) []*roaring.Bitmap {
	ra := roaring.NewBitmap()
	for v := range views {
		ra.Set(uint64(v.UnixMilli()))
	}
	rs := make([]*roaring.Bitmap, len(ts.views.ra))
	for i := range rs {
		rs[i] = ra.Clone()
	}
	ts.views.mu.RLock()
	for i := range rs {
		rs[i].And(ts.views.ra[i])
	}
	ts.views.mu.RUnlock()
	return rs
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
