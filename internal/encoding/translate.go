package encoding

import (
	"encoding/binary"

	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
)

func (e *Encoding) TranslateKey(field models.Field, value []byte) []byte {
	o := e.Allocate(6 + len(value))
	copy(o, keys.TranslateKeyPrefix)
	binary.BigEndian.PutUint32(o[2:], uint32(field))
	copy(o[6:], value)
	return o
}

func (e *Encoding) TranslateID(field models.Field, id uint64) []byte {
	o := e.Allocate(14)
	copy(o, keys.TranslateIDPrefix)
	binary.BigEndian.PutUint32(o[2:], uint32(field))
	binary.BigEndian.PutUint64(o[6:], id)
	return o
}

func (e *Encoding) TranslateSeq(field models.Field) []byte {
	o := e.Allocate(6)
	copy(o, keys.TranslateSeqPrefix)
	binary.BigEndian.PutUint32(o[2:], uint32(field))
	return o
}
