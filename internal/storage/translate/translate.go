package translate

import (
	"hash/maphash"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/storage/fields"
)

// Transtate global cache of translation per shard. We setmax shards to 256
type Transtate struct {
	shards [256][models.TranslatedFieldsSize]*mapping
	Seq    atomic.Uint64
}

func New(db *pebble.DB) (*Transtate, error) {
	t := new(Transtate)
	for i := range t.shards {
		for j := range models.TranslatedFieldsSize {
			t.shards[i][j] = &mapping{ma: make(map[uint64]uint64)}
		}
	}
	return t, t.Load(db)
}

// Load populates translation mapping.
func (t *Transtate) Load(db *pebble.DB) error {
	lo, hi := fields.MakeTranseIDRange()
	it, err := db.NewIter(&pebble.IterOptions{
		LowerBound: lo, UpperBound: hi,
	})
	if err != nil {
		return err
	}
	for it.First(); it.Valid(); it.Next() {
		field, shard, id := fields.BreakTranslationID(it.Key())
		ha := maphash.Bytes(seed, it.Value())
		t.shards[shard][models.AsMutex(field)].ma[ha] = id
	}
	return nil
}

func (t *Transtate) Get(field models.Field, shard uint64, value []byte) (uint64, bool) {
	return t.shards[shard][models.AsMutex(field)].Get(value)
}

func (t *Transtate) Find(field models.Field, shard uint64, value []byte) (uint64, bool) {
	return t.shards[shard][models.AsMutex(field)].Find(value)
}

var seed = maphash.MakeSeed()

type mapping struct {
	ma map[uint64]uint64
	mu sync.RWMutex
}

func (m *mapping) Get(value []byte) (id uint64, found bool) {
	ha := maphash.Bytes(seed, value)
	m.mu.RLock()
	id, found = m.ma[ha]
	m.mu.RUnlock()
	if found {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	id = uint64(len(m.ma) + 1)
	m.ma[ha] = id
	return
}

func (m *mapping) Find(value []byte) (id uint64, found bool) {
	ha := maphash.Bytes(seed, value)
	m.mu.RLock()
	id, found = m.ma[ha]
	m.mu.RUnlock()
	return
}
