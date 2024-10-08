package encoding

import (
	"encoding/binary"
	"slices"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/keys"
)

type Key struct {
	Time  uint64
	Shard uint32
	Field v1.Field
}

type Encoding struct {
	data []byte
}

func (e *Encoding) Reset() {
	clear(e.data)
	e.data = e.data[:0]
}

func (e *Encoding) Key(key Key) []byte {
	b := e.allocate(17)
	copy(b, keys.DataPrefix)
	binary.BigEndian.PutUint64(b[1:], key.Time)
	binary.BigEndian.PutUint32(b[9:], key.Shard)
	binary.BigEndian.PutUint32(b[13:], uint32(key.Field))
	return b
}

func (e *Encoding) Site(domain []byte) []byte {
	o := e.allocate(2 + len(domain))
	copy(o, keys.SitePrefix)
	copy(o[2:], domain)
	return o
}

func (e *Encoding) APIKeyName(key []byte) []byte {
	o := e.allocate(2 + len(key))
	copy(o, keys.APIKeyNamePrefix)
	copy(o[2:], key)
	return o
}

func (e *Encoding) APIKeyHash(hash []byte) []byte {
	o := e.allocate(2 + len(hash))
	copy(o, keys.APIKeyHashPrefix)
	copy(o[2:], hash)
	return o
}

func (e *Encoding) allocate(n int) []byte {
	off := len(e.data)
	e.data = slices.Grow(e.data, n)[:off+n]
	return e.data[off : off+n]
}
