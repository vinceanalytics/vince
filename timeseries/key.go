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

// Key identifies a key that stores a single aggregate value. Keys are
// lexicographically sorted. We use this property to arrange the key in such a
// way that it can be sorted by
//   - User
//   - Website
//   - Time
//   - Type of Aggregate
//
// We store aggregates in  Hour chunks. So Time refers to hours since unix epoch.
type Key [hashOffset + 4]byte

var metaKeyPool = &sync.Pool{
	New: func() any {
		var id Key
		return &id
	},
}

func (id *Key) SetAggregateType(u METRIC_TYPE) *Key {
	id[aggregateTypeOffset] = byte(u)
	return id
}

func (id *Key) SetProp(prop PROPS) *Key {
	id[propOffset] = byte(prop)
	return id
}

func (id *Key) SetUserID(u uint64) {
	binary.BigEndian.PutUint64(id[userOffset:], u)
}

func (id *Key) SetSiteID(u uint64) {
	binary.BigEndian.PutUint64(id[siteOffset:], u)
}

func (id *Key) HashU16(h uint16) []byte {
	binary.BigEndian.PutUint16(id[hashOffset:], h)
	return id[:][:hashOffset+2]
}

func (id *Key) HashU32(h uint32) *Key {
	binary.BigEndian.PutUint32(id[hashOffset:], h)
	return id
}

func (id *Key) Copy() *bytes.Buffer {
	b := smallBufferpool.Get().(*bytes.Buffer)
	b.Write(id[:])
	return b
}

func (id *Key) Prefix() []byte {
	return id[:hashOffset]
}

func (id *Key) String(s string) *bytes.Buffer {
	b := smallBufferpool.Get().(*bytes.Buffer)
	return id.StringBuffer(b, s)
}

func (id *Key) StringBuffer(b *bytes.Buffer, s string) *bytes.Buffer {
	b.Write(id[:])
	b.WriteString(s)
	return b
}

func (id *Key) GetUserID() uint64 {
	return binary.BigEndian.Uint64(id[userOffset:])
}

func (id *Key) GetSiteID() uint64 {
	return binary.BigEndian.Uint64(id[siteOffset:])
}

func (id *Key) Timestamp(ts time.Time) *Key {
	hours := timex.Timestamp(ts)
	binary.BigEndian.PutUint64(id[yearOffset:], hours)
	return id
}

func (id *Key) Release() {
	for i := 0; i < len(id); i += 1 {
		id[i] = 0
	}
	metaKeyPool.Put(id)
}

func newMetaKey() *Key {
	return metaKeyPool.Get().(*Key)
}
