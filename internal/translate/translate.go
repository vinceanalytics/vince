package translate

import (
	"hash/maphash"
	"sync"

	"github.com/vinceanalytics/vince/internal/models"
)

type Transtate struct {
	mapping [models.TranslatedFieldsSize]*mapping
}

func New() *Transtate {
	t := new(Transtate)
	for i := range t.mapping {
		t.mapping[i] = &mapping{ma: make(map[uint64]uint64)}
	}
	return t
}

func (t *Transtate) Get(field models.Field, value []byte) (uint64, bool) {
	return t.mapping[models.AsMutex(field)].Get(value)
}

var seed = maphash.MakeSeed()

type mapping struct {
	mu sync.RWMutex
	ma map[uint64]uint64
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
