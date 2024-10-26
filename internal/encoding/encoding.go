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
	b := make([]byte, 0, bitmapKeySize)
	b = BitmapBuf(shard, view, field, b)
	return b[:len(b):len(b)] // avoid passing around excess unused memory
}

func BitmapBuf(shard, view uint64, field models.Field, b []byte) []byte {
	b = append(b, keys.DataPrefix...)
	b = num(b, shard)
	b = num(b, view)
	b = num(b, uint64(field))
	return b
}

func Component(key []byte) (shard, view uint64) {
	shard = binary.BigEndian.Uint64(key[1:])
	view = binary.BigEndian.Uint64(key[1+8:])
	return
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

// Shard encodes key that stores bitmap of all views observed in shard. Uses
// keys.ShardsPrefix as prefix key and encodes shard in big endian notation.
func Shard(shard uint64) []byte {
	b := make([]byte, 0, 9)
	b = append(b, keys.ShardsPrefix...)
	return binary.BigEndian.AppendUint64(b, shard)
}

func num(b []byte, v uint64) []byte {
	return binary.BigEndian.AppendUint64(b, v)
}
