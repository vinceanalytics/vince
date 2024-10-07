package encoding

import (
	"encoding/binary"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/keys"
)

type Key struct {
	Time  uint64
	Shard uint32
	Field v1.Field
}

func EncodeKey(key Key) []byte {
	b := make([]byte, 17)
	copy(b, keys.DataPrefix)
	binary.BigEndian.PutUint64(b[1:], key.Time)
	binary.BigEndian.PutUint32(b[9:], key.Shard)
	binary.BigEndian.PutUint32(b[13:], uint32(key.Field))
	return b
}

func EncodeSite(key []byte) []byte {
	o := make([]byte, 2+len(key))
	copy(o, keys.SitePrefix)
	copy(o[2:], key)
	return o
}

func EncodeApiKeyName(key []byte) []byte {
	o := make([]byte, 2+len(key))
	copy(o, keys.APIKeyNamePrefix)
	copy(o[2:], key)
	return o
}

func EncodeApiKeyHash(hash []byte) []byte {
	o := make([]byte, 2+len(hash))
	copy(o, keys.APIKeyHashPrefix)
	copy(o[2:], hash)
	return o
}
