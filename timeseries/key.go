package timeseries

import (
	"encoding/binary"
	"sync"
	"time"
)

const (
	userOffset  = 0
	siteOffset  = 8
	yearOffset  = 16
	tableOffset = 18
	metaOffset  = 19
	hashOffset  = 20
)

type ID [19]byte

func (id *ID) SetTable(table byte) *ID {
	id[tableOffset] = byte(table)
	return id
}

func (id *ID) GetTable() byte {
	return id[tableOffset]
}

func (id *ID) SetUserID(u uint64) {
	binary.BigEndian.PutUint64(id[userOffset:], u)
}

func (id *ID) SetSiteID(u uint64) {
	binary.BigEndian.PutUint64(id[siteOffset:], u)
}

func (id *ID) GetUserID() uint64 {
	return binary.BigEndian.Uint64(id[userOffset:])
}

func (id *ID) GetSiteID() uint64 {
	return binary.BigEndian.Uint64(id[siteOffset:])
}

func (id *ID) Year(ts time.Time) *ID {
	binary.BigEndian.PutUint16(id[yearOffset:], uint16(ts.Year()))
	return id
}

func (id *ID) Release() {
	for i := 0; i < len(id); i += 1 {
		id[i] = 0
	}
	idBufPool.Put(id)
}

func (id *ID) Clone() *ID {
	x := newID()
	copy(x[:], id[:])
	return x
}

func newID() *ID {
	return idBufPool.Get().(*ID)
}

var idBufPool = &sync.Pool{
	New: func() any {
		var id ID
		return &id
	},
}
var metaKeyPool = &sync.Pool{
	New: func() any {
		var id MetaKey
		return &id
	},
}

// stores values for props
type MetaKey [28]byte

func (id *MetaKey) SetTable(table byte) *MetaKey {
	id[tableOffset] = byte(table)
	return id
}

func (id *MetaKey) SetMeta(table byte) *MetaKey {
	id[metaOffset] = byte(table)
	return id
}

func (id *MetaKey) GetTable() byte {
	return id[tableOffset]
}

func (id *MetaKey) SetUserID(u uint64) {
	binary.BigEndian.PutUint64(id[userOffset:], u)
}

func (id *MetaKey) SetSiteID(u uint64) {
	binary.BigEndian.PutUint64(id[siteOffset:], u)
}

func (id *MetaKey) HashU16(h uint16) []byte {
	binary.BigEndian.PutUint16(id[hashOffset:], h)
	return id[:][:hashOffset+2]
}

func (id *MetaKey) GetUserID() uint64 {
	return binary.BigEndian.Uint64(id[userOffset:])
}

func (id *MetaKey) GetSiteID() uint64 {
	return binary.BigEndian.Uint64(id[siteOffset:])
}

func (id *MetaKey) Year(ts time.Time) *MetaKey {
	binary.BigEndian.PutUint16(id[yearOffset:], uint16(ts.Year()))
	return id
}

func (id *MetaKey) Release() {
	for i := 0; i < len(id); i += 1 {
		id[i] = 0
	}
	metaKeyPool.Put(id)
}

func (id *MetaKey) Clone() *MetaKey {
	x := newMetaKey()
	copy(x[:], id[:])
	return x
}

func newMetaKey() *MetaKey {
	return metaKeyPool.Get().(*MetaKey)
}
