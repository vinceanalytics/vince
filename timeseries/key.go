package timeseries

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"

	"github.com/gernest/vince/pkg/timex"
)

const (
	userOffset   = 0
	siteOffset   = userOffset + 8
	metricOffset = siteOffset + 8
	propOffset   = metricOffset + 1
	yearOffset   = propOffset + 1
	keyOffset    = yearOffset + 8
)

// Key identifies a key that stores a single aggregate value. Keys are
// lexicographically sorted. We use this property to arrange the key in such a
// way that it can be sorted by
//   - User
//   - Website
//   - Metric (  visitors ... etc)
//   - Property ( event, page ... etc)
//   - Time
//   - Key Text
//
// We store aggregates in  Hour chunks. So Time refers to hours since unix epoch.
type Key [keyOffset]byte

var metaKeyPool = &sync.Pool{
	New: func() any {
		var id Key
		return &id
	},
}

func (id *Key) Metric(u Metric) *Key {
	id[metricOffset] = byte(u)
	return id
}

func (id *Key) Prop(prop Property) *Key {
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

func (id *Key) IndexBufferPrefix(b *bytes.Buffer, s string) *bytes.Buffer {
	b.Write(id[:yearOffset])
	b.WriteString(s)
	return b
}

// Converts index idx to a mike key. Mike keys ends with the text while index
// keys ends with timestamps.
func IndexToKey(idx []byte, o *bytes.Buffer) (mike *bytes.Buffer, text []byte, ts []byte) {
	mike = o
	// The last 4 bytes are for the timestamp
	ts = idx[len(idx)-4:]
	text = idx[yearOffset : len(idx)-4]
	o.Write(idx[:yearOffset])
	o.Write(ts)
	o.Write(text)
	return
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
