package encoding

import (
	"encoding/binary"
	"unsafe"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
)

type Resolution = v1.Resolution

const (
	Global = v1.Resolution_Global
	Minute = v1.Resolution_Minute
	Hour   = v1.Resolution_Hour
	Day    = v1.Resolution_Day
	Week   = v1.Resolution_Week
	Month  = v1.Resolution_Month
)

const (
	bmPrefix      = 0
	bmResolution  = bmPrefix + 1
	bmView        = bmResolution + 1
	bmDomain      = bmView + 8
	bmField       = bmDomain + 8
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

func (k *Key) WriteData(res Resolution, field models.Field, view, domainId, co uint64) {
	k[0] = keys.DataPrefix[0]
	k[bmResolution] = byte(res)
	binary.BigEndian.PutUint64(k[bmView:], view)
	binary.BigEndian.PutUint64(k[bmDomain:], domainId)
	k[bmField] = byte(field)
	binary.BigEndian.PutUint64(k[bmContainer:], co)
}

func (k *Key) WriteExistence(res Resolution, field models.Field, view, domainId, co uint64) {
	k[0] = keys.DataExistsPrefix[0]
	k[bmResolution] = byte(res)
	binary.BigEndian.PutUint64(k[bmView:], view)
	binary.BigEndian.PutUint64(k[bmDomain:], domainId)
	k[bmField] = byte(field)
	binary.BigEndian.PutUint64(k[bmContainer:], co)
}

func (k *Key) Component() (field models.Field, co uint64) {
	field = models.Field(k[bmField])
	co = binary.BigEndian.Uint64(k[bmContainer:])
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
	copy(o, keys.AcmePrefix)
	copy(o[2:], key)
	return o
}

func num(b []byte, v uint64) []byte {
	return binary.BigEndian.AppendUint64(b, v)
}
