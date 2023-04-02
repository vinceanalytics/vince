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
)

type ID [20]byte

func (id *ID) SetTable(table byte) *ID {
	id[tableOffset] = byte(table)
	return id
}

func (id *ID) SetMeta(table byte) *ID {
	id[metaOffset] = byte(table)
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
	x := NewID()
	copy(x[:], id[:])
	return x
}

func NewID() *ID {
	return idBufPool.Get().(*ID)
}

var idBufPool = &sync.Pool{
	New: func() any {
		var id ID
		return &id
	},
}
