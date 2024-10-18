package ro2

import (
	"encoding/binary"
	"sync"

	"github.com/dgryski/go-farm"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"

	"github.com/dgraph-io/badger/v4/y"
)

const (
	// MaxLeaseSize determines number of UID assigned from memory pefore persisting
	// on the database.
	MaxLeaseSize = 8 << 10
)

func (o *Store) ID(field models.Field, key []byte) (id uint64) {
	e := get()
	k := e.TranslateKey(field, key)
	hash := farm.Fingerprint64(k)
	o.mu.RLock()
	id = o.tree.Get(hash)
	o.mu.RUnlock()
	release(e)
	return
}

func (m *Store) AssignUid(field models.Field, xid []byte) uint64 {
	e := get()
	defer release(e)
	key := e.TranslateKey(field, xid)
	hash := farm.Fingerprint64(key)
	m.mu.RLock()
	uid := m.tree.Get(hash)
	m.mu.RUnlock()
	if uid > 0 {
		return uid
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	idx := field.TranslateIndex()
	newUID, err := m.ranges[idx].Next()
	y.Check(err)
	if newUID == 0 {
		newUID, err = m.ranges[idx].Next()
		y.Check(err)
	}

	m.tree.Set(hash, newUID)
	m.keys[idx] = append(m.keys[idx], xid)
	m.values[idx] = append(m.values[idx], newUID)
	return newUID
}

func (m *Store) Flush() error {
	b := m.db.NewWriteBatch()
	e := get()
	defer func() {
		// avoid keeping large buffers in the bool
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
		f := models.TranslateIndex(i)

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
