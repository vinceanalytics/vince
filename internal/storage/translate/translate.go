package translate

import (
	"hash/maphash"
	"sync"

	"github.com/vinceanalytics/vince/internal/models"
)

// Transtate global cache of translation per shard. We setmax shards to 256
type Transtate struct {
	shards [256][models.TranslatedFieldsSize]*mapping
}

func New() *Transtate {
	t := new(Transtate)
	for i := range t.shards {
		for j := range models.TranslatedFieldsSize {
			t.shards[i][j] = &mapping{ma: make(map[uint64]uint64)}
		}
	}
	return t
}

func (t *Transtate) Get(field models.Field, shard uint64, value []byte) (uint64, bool) {
	return t.shards[shard][models.AsMutex(field)].Get(value)
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
