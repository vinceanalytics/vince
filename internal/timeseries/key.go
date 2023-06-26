package timeseries

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

const (
	userOffset   = 0
	siteOffset   = userOffset + 8
	metricOffset = siteOffset + 8
	propOffset   = metricOffset + 1
	keyOffset    = propOffset + 1
)

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

var zero = make([]byte, 8)

func (id *Key) clone() *Key {
	o := newMetaKey()
	copy(o[:], id[:])
	return o
}

func (id *Key) uid(u, s uint64) *Key {
	binary.BigEndian.PutUint64(id[userOffset:], u)
	binary.BigEndian.PutUint64(id[siteOffset:], s)
	return id
}

func DebugKey(id []byte) string {
	uid := binary.BigEndian.Uint64(id[userOffset:])
	sid := binary.BigEndian.Uint64(id[userOffset:])
	metric := Metric(id[metricOffset])
	prop := Property(id[propOffset])
	key := id[keyOffset : len(id)-6]
	ts := Time(id[len(id)-6:])
	g := smallBufferpool.Get().(*bytes.Buffer)
	fmt.Fprintf(g, "/%s/%d/%d/%s/%s/%s", ts.Format(time.DateTime), uid, sid, metric, prop, string(key))
	o := g.String()
	g.Reset()
	smallBufferpool.Put(g)
	return o
}

func DebugPrefix(id []byte) string {
	uid := binary.BigEndian.Uint64(id[userOffset:])
	sid := binary.BigEndian.Uint64(id[userOffset:])
	metric := Metric(id[metricOffset])
	prop := Property(id[propOffset])
	key := id[keyOffset:]
	g := smallBufferpool.Get().(*bytes.Buffer)
	fmt.Fprintf(g, "/%d/%d/%s/%s/%s", uid, sid, metric, prop, string(key))
	o := g.String()
	g.Reset()
	smallBufferpool.Put(g)
	return o
}

func setTs(b []byte, ms uint64) {
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
}

func Time(id []byte) time.Time {
	ms := uint64(id[5]) | uint64(id[4])<<8 |
		uint64(id[3])<<16 | uint64(id[2])<<24 |
		uint64(id[1])<<32 | uint64(id[0])<<40
	s := int64(ms / 1e3)
	ns := int64((ms % 1e3) * 1e6)
	return time.Unix(s, ns)
}

func (id *Key) key(ts []byte, s string, ls *txnBufferList) *bytes.Buffer {
	k := ls.Get()
	k.Write(id[:])
	k.WriteString(s)
	k.Write(ts)
	return k
}

func (id *Key) idx(b *bytes.Buffer, s string) *bytes.Buffer {
	b.Write(id[:])
	b.WriteString(s)
	return b
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
