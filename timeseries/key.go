package timeseries

import (
	"bytes"
	"encoding/binary"
	"sync"
)

const (
	userOffset   = 0
	siteOffset   = userOffset + 8
	metricOffset = siteOffset + 8
	propOffset   = metricOffset + 1
	timeOffset   = propOffset + 1
	keyOffset    = timeOffset + 6
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
// Time is in milliseconds since unix epoch truncated to the nearest Hour.
type Key [keyOffset]byte

var metaKeyPool = &sync.Pool{
	New: func() any {
		var id Key
		return &id
	},
}

func (id *Key) metric(u Metric) *Key {
	id[metricOffset] = byte(u)
	return id
}

func (id *Key) prop(p Property) *Key {
	id[propOffset] = byte(p)
	return id
}

func (id *Key) uid(u uint64) {
	binary.BigEndian.PutUint64(id[userOffset:], u)
}

func (id *Key) sid(s uint64) {
	binary.BigEndian.PutUint64(id[siteOffset:], s)
}

func (id *Key) ts(ms uint64) *Key {
	(*id)[timeOffset+0] = byte(ms >> 40)
	(*id)[timeOffset+1] = byte(ms >> 32)
	(*id)[timeOffset+2] = byte(ms >> 24)
	(*id)[timeOffset+3] = byte(ms >> 16)
	(*id)[timeOffset+4] = byte(ms >> 8)
	(*id)[timeOffset+5] = byte(ms)
	return id
}

func Time(id []byte) uint64 {
	return uint64(id[5]) | uint64(id[4])<<8 |
		uint64(id[3])<<16 | uint64(id[2])<<24 |
		uint64(id[1])<<32 | uint64(id[0])<<40
}

type IDToSave struct {
	mike  *bytes.Buffer
	index *bytes.Buffer
}

func (id *Key) key(s string, ls *txnBufferList) *IDToSave {

	k := ls.Get()
	k.Write(id[:])
	k.WriteString(s)

	// Index [ Txt/Time ]
	idx := ls.Get()
	idx.Write(id[:timeOffset])
	idx.WriteString(s)
	idx.Write(id[timeOffset:])

	return &IDToSave{
		mike:  k,
		index: idx,
	}
}

func (id *Key) IndexBufferPrefix(b *bytes.Buffer, s string) *bytes.Buffer {
	b.Write(id[:timeOffset])
	b.WriteString(s)
	return b
}

// Converts index idx to a mike key. Mike keys ends with the text while index
// keys ends with timestamps.
func IndexToKey(idx []byte, o *bytes.Buffer) (mike *bytes.Buffer, text []byte, ts []byte) {
	mike = o
	// The last 6 bytes are for the timestamp
	ts = idx[len(idx)-6:]
	text = idx[timeOffset : len(idx)-6]
	o.Write(idx[:timeOffset])
	o.Write(ts)
	o.Write(text)
	return
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
