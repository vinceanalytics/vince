package timeseries

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"

	"github.com/gernest/vince/pkg/timex"
)

const (
	userOffset          = 0
	siteOffset          = userOffset + 8
	aggregateTypeOffset = siteOffset + 8
	propOffset          = aggregateTypeOffset + 1
	yearOffset          = propOffset + 1
	hashOffset          = yearOffset + 8
)

// ID identifies a key that stores a single aggregate value. Keys are
// lexicographically sorted. We use this property to arrange the key in such a
// way that it can be sorted by
//   - User
//   - Website
//   - Time
//   - Type of Aggregate
//
// We store aggregates in  Hour chunks. So Time refers to hours since unix epoch.
type ID [hashOffset]byte

func (id *ID) SetUserID(u uint64) {
	binary.BigEndian.PutUint64(id[userOffset:], u)
}

func (id *ID) SetSiteID(u uint64) {
	binary.BigEndian.PutUint64(id[siteOffset:], u)
}

func (id *ID) SetAggregateType(u METRIC_TYPE) *ID {
	id[aggregateTypeOffset] = byte(u)
	return id
}

func (id *ID) GetUserID() uint64 {
	return binary.BigEndian.Uint64(id[userOffset:])
}

func (id *ID) GetSiteID() uint64 {
	return binary.BigEndian.Uint64(id[siteOffset:])
}

func (id *ID) SitePrefix() []byte {
	return id[:yearOffset]
}

// Timestamp hours since unix epoch
func (id *ID) Timestamp(ts time.Time) *ID {
	hours := timex.Timestamp(ts)
	binary.BigEndian.PutUint64(id[yearOffset:], hours)
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
type MetaKey [hashOffset + 4]byte

func (id *MetaKey) SetAggregateType(u METRIC_TYPE) *MetaKey {
	id[aggregateTypeOffset] = byte(u)
	return id
}

func (id *MetaKey) SetProp(table byte) *MetaKey {
	id[propOffset] = byte(table)
	return id
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

func (id *MetaKey) HashU32(h uint32) *MetaKey {
	binary.BigEndian.PutUint32(id[hashOffset:], h)
	return id
}

func (id *MetaKey) Copy() *bytes.Buffer {
	b := smallBufferpool.Get().(*bytes.Buffer)
	b.Write(id[:])
	return b
}

func (id *MetaKey) Prefix() []byte {
	return id[:hashOffset]
}

func (id *MetaKey) String(s string) *bytes.Buffer {
	b := smallBufferpool.Get().(*bytes.Buffer)
	return id.StringBuffer(b, s)
}

func (id *MetaKey) StringBuffer(b *bytes.Buffer, s string) *bytes.Buffer {
	b.Write(id[:])
	b.WriteString(s)
	return b
}

func (id *MetaKey) GetUserID() uint64 {
	return binary.BigEndian.Uint64(id[userOffset:])
}

func (id *MetaKey) GetSiteID() uint64 {
	return binary.BigEndian.Uint64(id[siteOffset:])
}

func (id *MetaKey) Timestamp(ts time.Time) *MetaKey {
	hours := timex.Timestamp(ts)
	binary.BigEndian.PutUint64(id[yearOffset:], hours)
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
