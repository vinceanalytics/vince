package keys

import (
	"bytes"
	"encoding/base32"
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
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
	if k == nil {
		return
	}
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
	return New(v1.StorePrefix_BLOCK_METADATA).Path(
		domain, uid,
	)
}

// Returns key which stores a block index in badger db.
func BlockIndex(domain, uid string) *Key {
	return New(v1.StorePrefix_BLOCK_INDEX).Path(
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

func Cluster() *Key {
	return New(v1.StorePrefix_CLUSTER)
}

func AClient(id string) *Key {
	return New(v1.StorePrefix_OAUTH_CLIENT).Path(id)
}

func AAccess(id string) *Key {
	return New(v1.StorePrefix_OAUTH_ACCESS).Path(id)
}

func AAuthorize(id string) *Key {
	return New(v1.StorePrefix_OAUTH_AUTHORIZE).Path(id)
}

func ARefresh(id string) *Key {
	return New(v1.StorePrefix_OAUTH_REFRESH).Path(id)
}

func Snippet(uid, sid string) *Key {
	return New(v1.StorePrefix_SNIPPET).Path(uid, sid)
}
