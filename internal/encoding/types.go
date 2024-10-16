package encoding

import (
	"encoding/binary"
	"slices"

	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
)

type Key struct {
	Time  uint64
	Field models.Field
}

type Encoding struct {
	data []byte
}

func (e *Encoding) Reset() {
	clear(e.data)
	e.data = e.data[:0]
}

// Like Reset but releases excess pacapity if any
func (e *Encoding) Clip(sz int) {
	clear(e.data)
	if len(e.data) > sz {
		e.data = e.data[:sz:sz]
	}
	e.data = e.data[:0]
}

const BitmapKeySize = 1 + //prefix
	8 + //shard
	8 + // timestamp
	1 + // field
	1 // index

func (e *Encoding) Bitmap(ts, shard uint64, field models.Field, index byte) []byte {
	return Bitmap(ts, shard, field, index, e.Allocate(BitmapKeySize))
}

func Bitmap(ts, shard uint64, field models.Field, index byte, buf []byte) []byte {
	b := buf
	copy(b, keys.DataPrefix)
	binary.BigEndian.PutUint64(b[1:], shard)
	binary.BigEndian.PutUint64(b[1+8:], ts)
	b[1+8+8] = byte(field)
	b[1+8+8+1] = index
	return b
}

func (e *Encoding) Site(domain []byte) []byte {
	o := e.Allocate(2 + len(domain))
	copy(o, keys.SitePrefix)
	copy(o[2:], domain)
	return o
}

func (e *Encoding) APIKeyName(key []byte) []byte {
	o := e.Allocate(2 + len(key))
	copy(o, keys.APIKeyNamePrefix)
	copy(o[2:], key)
	return o
}

func (e *Encoding) APIKeyHash(hash []byte) []byte {
	o := e.Allocate(2 + len(hash))
	copy(o, keys.APIKeyHashPrefix)
	copy(o[2:], hash)
	return o
}

func (e *Encoding) Allocate(n int) []byte {
	e.Grow(n)
	off := len(e.data)
	e.data = e.data[:off+n]
	return e.data[off : off+n]
}

func (e *Encoding) Grow(n int) {
	if len(e.data)+n < cap(e.data) {
		return
	}
	// Calculate new capacity.
	growBy := len(e.data) + n

	// Don't allocate more than 1GB at a time.
	if growBy > 1<<30 {
		growBy = 1 << 30
	}
	// Allocate at least n, even if it exceeds the 1GB limit above.
	if n > growBy {
		growBy = n
	}
	e.data = slices.Grow(e.data, growBy)
}
