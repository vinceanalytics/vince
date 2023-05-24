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
	keyOffset           = yearOffset + 8
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
type Key [keyOffset]byte

var metaKeyPool = &sync.Pool{
	New: func() any {
		var id Key
		return &id
	},
}

func (id *Key) SetAggregateType(u Metric) *Key {
	id[aggregateTypeOffset] = byte(u)
	return id
}

func (id *Key) SetProp(prop Property) *Key {
	id[propOffset] = byte(prop)
	return id
}

func (id *Key) SetUserID(u uint64) {
	binary.BigEndian.PutUint64(id[userOffset:], u)
}

func (id *Key) SetSiteID(u uint64) {
	binary.BigEndian.PutUint64(id[siteOffset:], u)
}

func (id *Key) Copy() *bytes.Buffer {
	b := smallBufferpool.Get().(*bytes.Buffer)
	b.Write(id[:])
	return b
}

type IDToSave struct {
	mike  *bytes.Buffer
	index *bytes.Buffer
}

func (id *Key) Key(s string) *IDToSave {
	b := smallBufferpool.Get().(*bytes.Buffer)
	idx := smallBufferpool.Get().(*bytes.Buffer)
	return &IDToSave{
		mike:  id.KeyBuffer(b, s),
		index: id.IndexBuffer(idx, s),
	}
}

func (id *Key) KeyBuffer(b *bytes.Buffer, s string) *bytes.Buffer {
	b.Write(id[:])
	b.WriteString(s)
	return b
}

// IndexBuffer returns a key with timestamp as a suffix. This allows for faster
// key lookup without the knowledge of the timestamp.
func (id *Key) IndexBuffer(b *bytes.Buffer, s string) *bytes.Buffer {
	b.Write(id[:yearOffset])
	b.WriteString(s)
	b.Write(id[yearOffset:])
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
