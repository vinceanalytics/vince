package encoding

import (
	"encoding/binary"
	"time"

	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
)

type Resolution byte

const (
	Global Resolution = iota
	Minute
	Hour
	Day
	Week
	Month
)

const (
	bmPrefix      = 0
	bmResolution  = bmPrefix + 1
	bmView        = bmResolution + 1
	bmField       = bmView + 8
	bmShard       = bmField + 1
	bitmapKeySize = bmShard + 8
)

type viewFn func(time.Time) uint64

func toView(cmp func(time.Time) time.Time) viewFn {
	return func(t time.Time) uint64 {
		return uint64(cmp(t).UnixMilli())
	}
}

func Bitmap(field models.Field, res Resolution, ts uint64, shard uint64) []byte {
	b := make([]byte, 0, bitmapKeySize)
	return BitmapBuf(field, res, ts, shard, b)
}

func BitmapBuf(field models.Field, res Resolution, ts uint64, shard uint64, b []byte) []byte {
	b = append(b, keys.DataPrefix...)
	b = append(b, byte(res))
	b = num(b, ts)
	b = append(b, byte(field))
	b = num(b, shard)
	return b
}

func Min(field models.Field, res Resolution, ts uint64) []byte {
	return Bitmap(field, res, ts, 0)
}

func Max(field models.Field, res Resolution, ts uint64) []byte {
	return Bitmap(field, res, ts, 0)
}

func Component(key []byte) (field models.Field, res Resolution, view, shard uint64) {
	field = models.Field(key[bmField])
	res = Resolution(key[bmResolution])
	view = binary.BigEndian.Uint64(key[bmView:])
	shard = binary.BigEndian.Uint64(key[bmShard:])
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

func num(b []byte, v uint64) []byte {
	return binary.BigEndian.AppendUint64(b, v)
}
