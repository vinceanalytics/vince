package store

import (
	"encoding/binary"
	"slices"
	"sync"

	"github.com/dgryski/go-farm"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"

	"github.com/dgraph-io/badger/v4/y"
)

const (
	// maxTranslationKeyLeaseSize determines number of UID assigned from memory pefore persisting
	// on the database.
	maxTranslationKeyLeaseSize = 8 << 10
)

func (m *Store) ID(field models.Field, xid []byte) (id uint64) {
	e := get()
	defer release(e)

	key := e.TranslateKey(field, xid)

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.trie.Get(key)
}

func (m *Store) Search(field models.Field, prefix []byte, f func(key []byte, val uint64)) error {
	e := get()
	defer release(e)

	m.mu.RLock()
	defer m.mu.RUnlock()

	// we expand capacity of prefix to avoid extra allocations for smaller key,
	// 1kb is reasonbale buffer size.
	prefix = slices.Grow(e.TranslateKey(field, prefix), 1<<10)
	return m.trie.IterateWIthPrefix(prefix, func(key []byte, uid uint64) error {
		// remove field prefix before calling f
		f(key[3:], uid)
		return nil
	})
}

func (m *Store) AssignUid(field models.Field, xid []byte) uint64 {
	key := m.enc.TranslateKey(field, xid)
	hash := farm.Fingerprint64(key)
	uid := m.tree.Get(hash)
	if uid > 0 {
		return uid
	}

	idx := field.Mutex()
	newUID, err := m.ranges[idx].Next()
	y.Check(err)
	if newUID == 0 {
		newUID, err = m.ranges[idx].Next()
		y.Check(err)
	}

	m.tree.Set(hash, newUID)
	m.keys[idx] = append(m.keys[idx], xid)
	m.values[idx] = append(m.values[idx], newUID)
	m.mu.Lock()
	m.trie.Put(key, newUID)
	m.mu.Unlock()
	return newUID
}

func (m *Store) Flush() error {
	b := m.db.NewWriteBatch()
	e := get()
	defer func() {
		// avoid keeping large buffers in the pool
		e.Clip(4 << 10)
		encPool.Put(e)
		for i := range m.keys {
			clear(m.keys[i])
			m.keys[i] = m.keys[i][:0]
			clear(m.values[i])
			m.values[i] = m.values[i][:0]
		}
	}()

	for i := range m.keys {
		f := models.Mutex(i)

		for j := range m.keys[i] {
			value := e.Allocate(8)
			binary.BigEndian.PutUint64(value, m.values[i][j])
			err := b.Set(
				e.TranslateKey(f, m.keys[i][j]),
				value,
			)
			y.Check(err)
			err = b.Set(e.TranslateID(f, m.values[i][j]), m.keys[i][j])
			y.Check(err)
		}
	}
	return b.Flush()
}

var encPool = &sync.Pool{New: func() any { return new(encoding.Encoding) }}

func get() *encoding.Encoding {
	return encPool.Get().(*encoding.Encoding)
}

func release(e *encoding.Encoding) {
	e.Reset()
	encPool.Put(e)
}
