package timeseries

import (
	"context"
	"sync"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/util/data"
	xt "github.com/vinceanalytics/vince/internal/util/translation"
	"github.com/vinceanalytics/vince/internal/util/trie"
)

type Timeseries struct {
	db *pebble.DB
	ba *batch

	mu   sync.RWMutex
	trie *trie.Trie
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
	return ts.ba.save()
}

func (ts *Timeseries) Add(m *models.Model) error {
	return ts.ba.add(m)
}

func (ts *Timeseries) NewBitmap(ctx context.Context, shard uint64, view uint64, field models.Field) (b *roaring.Bitmap) {
	buf := make([]byte, encoding.BitmapKeySize)
	data.Get(ts.db, encoding.Bitmap(shard, view, field, buf), func(val []byte) error {
		b = roaring.FromBufferWithCopy(val)
		return nil
	})
	return
}
