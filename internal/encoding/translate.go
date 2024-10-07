package encoding

import (
	"encoding/binary"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/keys"
)

func EncodeTranslateKey(field v1.Field, value string) []byte {
	o := make([]byte, 6+len(value))
	copy(o, keys.TranslateKeyPrefix)
	binary.BigEndian.PutUint32(o[2:], uint32(field))
	copy(o[6:], []byte(value))
	return o
}

func EncodeTranslateID(field v1.Field, id uint64) []byte {
	o := make([]byte, 2+4+8)
	copy(o, keys.TranslateIDPrefix)
	binary.BigEndian.PutUint32(o[2:], uint32(field))
	binary.BigEndian.PutUint64(o[6:], id)
	return o
}

func EncodeTranslateSeq(field v1.Field) []byte {
	o := make([]byte, 6)
	copy(o, keys.TranslateSeqPrefix)
	binary.BigEndian.PutUint32(o[2:], uint32(field))
	return o
}
