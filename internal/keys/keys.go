package keys

import (
	"bytes"
	"encoding/base32"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
)

const DefaultNS = "vince"

func Path(prefix v1.StorePrefix, parts ...string) []byte {
	var b bytes.Buffer
	b.WriteString(DefaultNS)
	b.WriteByte('/')
	b.WriteString(prefix.String())
	for i := range parts {
		b.WriteByte('/')
		b.WriteString(parts[i])
	}
	return b.Bytes()
}

// Returns a key that stores Site object in the badger database with the given
// domain.
func Site(domain string) []byte {
	return Path(v1.StorePrefix_SITES, domain)
}

// Returns key which stores a block index in badger db.
func BlockMetadata(domain, uid string) []byte {
	return Path(v1.StorePrefix_BLOCK_METADATA,
		domain, uid,
	)
}

func BlockIndex(domain, uid string, col v1.Column) []byte {
	return Path(v1.StorePrefix_BLOCK_INDEX,
		domain, uid, col.String(),
	)
}

// Returns a key which stores account object in badger db.
func Account(name string) []byte {
	return Path(v1.StorePrefix_ACCOUNT, name)
}

func Token(token string) []byte {
	return Path(v1.StorePrefix_TOKEN, base32.StdEncoding.EncodeToString([]byte(token)))
}

func Cluster() []byte {
	return Path(v1.StorePrefix_CLUSTER)
}

func AClient(id string) []byte {
	return Path(v1.StorePrefix_OAUTH_CLIENT, id)
}

func AAccess(id string) []byte {
	return Path(v1.StorePrefix_OAUTH_ACCESS, id)
}

func AAuthorize(id string) []byte {
	return Path(v1.StorePrefix_OAUTH_AUTHORIZE, id)
}

func ARefresh(id string) []byte {
	return Path(v1.StorePrefix_OAUTH_REFRESH, id)
}

func Snippet(sid string) []byte {
	return Path(v1.StorePrefix_SNIPPET, sid)
}

func Import(name string) []byte {
	return Path(v1.StorePrefix_IMPORT, name)
}
