package translate

import (
	"bytes"
	"encoding/binary"
	"hash/maphash"
	"math"
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
	it, err := db.NewIter(&pebble.IterOptions{})
	if err != nil {
		return err
	}
	lo := fields.MakeTranslationID(0, 0, 0)
	hi := fields.MakeTranslationID(0, 0, 0)

	for shard := range t.shards {
		binary.BigEndian.PutUint64(lo[fields.ShardOffset:], uint64(shard))
		binary.BigEndian.PutUint64(hi[fields.ShardOffset:], uint64(shard))
		ma := t.shards[shard]
		var hasData bool
		for f := range models.TranslatedFieldsSize {
			field := models.Mutex(int(f))
			lo[fields.FieldOffset] = byte(field)
			hi[fields.FieldOffset] = byte(field)
			fx := ma[field]
			binary.BigEndian.PutUint64(hi[fields.TranslationIDOffset:], math.MaxUint64)
			for it.SeekGE(lo); it.Valid() && bytes.Compare(it.Key(), hi) == -1; it.Next() {
				if !hasData {
					hasData = true
				}
				id := binary.BigEndian.Uint64(it.Key()[fields.TranslationIDOffset:])
				hash := maphash.Bytes(seed, it.Value())
				fx.ma[hash] = id
			}
		}
		if !it.Valid() {
			break
		}
	}

	seqKey := fields.MakeSeqKey()
	if it.SeekGE(seqKey) && bytes.Equal(it.Key(), seqKey) {
		t.Seq.Store(binary.BigEndian.Uint64(it.Value()))
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
