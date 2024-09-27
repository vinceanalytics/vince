package encoding

import (
	"encoding/binary"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

func EncodeTranslateKey(field v1.Field, value string) []byte {
	o := make([]byte, 4+len(value))
	binary.BigEndian.PutUint32(o, uint32(field))
	return append(o[:4], []byte(value)...)
}

func EncodeTranslateID(field v1.Field, id uint64) []byte {
	o := make([]byte, 4+8)
	binary.BigEndian.PutUint32(o, uint32(field))
	binary.BigEndian.PutUint64(o[4:], id)
	return o
}

func EncodeTranslateSeq(field v1.Field) []byte {
	o := make([]byte, 4)
	binary.BigEndian.PutUint32(o, uint32(field))
	return o
}
