package encoding

import (
	"encoding/binary"
	"time"
	"unsafe"

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
	bmField       = bmPrefix + 1
	bmContainer   = bmField + 1
	BitmapKeySize = bmContainer + 8
)

type Key [BitmapKeySize]byte

func (k *Key) Reset() {
	clear(k[:])
}

func (k *Key) Bytes() []byte {
	return k[:]
}

func From(a []byte) *Key {
	return (*Key)(unsafe.Pointer(&a[0]))
}

func (k *Key) Write(field models.Field, co uint64) {
	k[0] = keys.DataPrefix[0]
	k[bmField] = byte(field)
	binary.BigEndian.PutUint64(k[bmContainer:], co)
}

func (k *Key) Component() (field models.Field, co uint64) {
	field = models.Field(k[bmField])
	co = binary.BigEndian.Uint64(k[bmContainer:])
	return
}

type viewFn func(time.Time) uint64

func toView(cmp func(time.Time) time.Time) viewFn {
	return func(t time.Time) uint64 {
		return uint64(cmp(t).UnixMilli())
	}
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
