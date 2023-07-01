package timeseries

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/vinceanalytics/vince/pkg/spec"
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

func (id *Key) metric(u spec.Metric) *Key {
	id[metricOffset] = byte(u)
	return id
}

func (id *Key) prop(p spec.Property) *Key {
	id[propOffset] = byte(p)
	return id
}

var zero = make([]byte, 8)

func (id *Key) String() string {
	return formatID(id[:])
}

func formatID(id []byte) string {
	return fmt.Sprintf(
		"%d/%d/%s/%s",
		binary.BigEndian.Uint64(id[userOffset:]),
		binary.BigEndian.Uint64(id[siteOffset:]),
		spec.Metric(id[metricOffset]),
		spec.Property(id[propOffset]),
	)
}

func (id *Key) uid(u, s uint64) *Key {
	binary.BigEndian.PutUint64(id[userOffset:], u)
	binary.BigEndian.PutUint64(id[siteOffset:], s)
	return id
}

func DebugKey(id []byte) string {
	uid := binary.BigEndian.Uint64(id[userOffset:])
	sid := binary.BigEndian.Uint64(id[siteOffset:])
	metric := spec.Metric(id[metricOffset])
	prop := spec.Property(id[propOffset])
	key := id[keyOffset : len(id)-8]
	ts := Time(id[len(id)-8:])
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
	metric := spec.Metric(id[metricOffset])
	prop := spec.Property(id[propOffset])
	key := id[keyOffset:]
	g := smallBufferpool.Get().(*bytes.Buffer)
	fmt.Fprintf(g, "/%d/%d/%s/%s/%s", uid, sid, metric, prop, string(key))
	o := g.String()
	g.Reset()
	smallBufferpool.Put(g)
	return o
}

func Time(id []byte) time.Time {
	return time.UnixMilli(
		int64(binary.BigEndian.Uint64(id)),
	)
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
