package keys

import (
	"bytes"
	"encoding/base32"
	"strconv"
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/v1"
)

type Key struct {
	b      bytes.Buffer
	ns     string
	prefix v1.StorePrefix
}

var pool = &sync.Pool{New: func() any { return new(Key) }}

const DefaultNS = "vince"

func New(prefix v1.StorePrefix, ns ...string) *Key {
	k := pool.Get().(*Key)
	k.prefix = prefix
	if len(ns) > 0 {
		k.ns = ns[0]
	} else {
		k.ns = DefaultNS
	}
	return k
}

func (k *Key) Path(parts ...string) *Key {
	k.b.WriteString(k.ns)
	k.b.WriteByte('/')
	k.b.WriteString(k.prefix.String())
	for i := range parts {
		k.b.WriteByte('/')
		k.b.WriteString(parts[i])
	}
	return k
}

func (k *Key) Release() {
	k.b.Reset()
	pool.Put(k)
}

func (k *Key) Bytes() []byte {
	return k.b.Bytes()
}

// Returns a key that stores Site object in the badger database with the given
// domain.
func Site(domain string) *Key {
	return New(v1.StorePrefix_SITES).Path(domain)
}

// Returns key which stores a block metadata in badger db.
func BlockMeta(domain, uid string) *Key {
	return New(v1.StorePrefix_BLOCKS).Path(
		v1.Block_Key_METADATA.String(),
		domain, uid,
	)
}

// Returns key which stores a block index in badger db.
func BlockIndex(domain, uid string) *Key {
	return New(v1.StorePrefix_BLOCKS).Path(
		v1.Block_Key_INDEX.String(),
		domain, uid,
	)
}

// Returns a key which stores account object in badger db.
func Account(name string) *Key {
	return New(v1.StorePrefix_ACCOUNT).Path(name)
}

func Token(token string) *Key {
	return New(v1.StorePrefix_TOKEN).Path(base32.StdEncoding.EncodeToString([]byte(token)))
}

func RaftLog(id int64) *Key {
	if id == -1 {
		return New(v1.StorePrefix_RAFT_LOGS).Path("")
	}
	return New(v1.StorePrefix_RAFT_LOGS).Path(strconv.FormatInt(id, 10))
}

func RaftStable(key []byte) *Key {
	return New(v1.StorePrefix_RAFT_STABLE).Path(base32.StdEncoding.EncodeToString(key))
}

func RaftSnapshotData(id string) *Key {
	return New(v1.StorePrefix_RAFT_SNAPSHOT).Path(
		v1.Raft_Snapshot_Key_DATA.String(), id,
	)
}

func RaftSnapshotMeta(id string) *Key {
	return New(v1.StorePrefix_RAFT_SNAPSHOT).Path(
		v1.Raft_Snapshot_Key_META.String(), id,
	)
}

func Cluster() *Key {
	return New(v1.StorePrefix_CLUSTER)
}
