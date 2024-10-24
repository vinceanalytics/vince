package encoding

import (
	"encoding/binary"

	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
)

const bitmapKeySize = 1 + //prefix
	8 + //shard
	8 + // view
	1 // field

func Bitmap(shard, view uint64, field models.Field) []byte {
	// we take advantage of protobuf variable length encoding to create compact
	// key space
	//
	// This ensures we have unlimited shards and view encoding to allow scalling
	// to billions of events.
	b := make([]byte, 0, bitmapKeySize)
	b = BitmapBuf(shard, view, field, b)
	return b[:len(b):len(b)] // avoid passing around excess unused memory
}

func BitmapBuf(shard, view uint64, field models.Field, b []byte) []byte {
	// we take advantage of protobuf variable length encoding to create compact
	// key space
	//
	// This ensures we have unlimited shards and view encoding to allow scalling
	// to billions of events.
	b = append(b, keys.DataPrefix...)
	b = num(b, shard)
	b = num(b, view)
	b = num(b, uint64(field))
	return b
}

func Site(domain []byte) []byte {
	o := make([]byte, 2+len(domain))
	copy(o, keys.SitePrefix)
	copy(o[2:], domain)
	return o
}

func APIKeyName(key []byte) []byte {
	o := make([]byte, 2+len(key))
	copy(o, keys.APIKeyNamePrefix)
	copy(o[2:], key)
	return o
}

func APIKeyHash(hash []byte) []byte {
	o := make([]byte, 2+len(hash))
	copy(o, keys.APIKeyHashPrefix)
	copy(o[2:], hash)
	return o
}

func ACME(key []byte) []byte {
	o := make([]byte, 2+len(key))
	copy(o, keys.APIKeyHashPrefix)
	copy(o[2:], key)
	return o
}

func num(b []byte, v uint64) []byte {
	b = appendVarint(b, v)
	return b // add separator
}

// AppendVarint appends v to b as a varint-encoded uint64.
func appendVarint(b []byte, v uint64) []byte {
	return binary.AppendUvarint(b, v)
}
