package encoding

import (
	"encoding/binary"
	"fmt"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

type Key struct {
	Time  uint64
	Shard uint32
	Field v1.Field
}

func (k *Key) String() string {
	return fmt.Sprintf("%d/%d/%s", k.Time, k.Shard, k.Field)
}

func EncodeKey(key Key) []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[:8], key.Time)
	binary.BigEndian.PutUint32(b[8:], key.Shard)
	binary.BigEndian.PutUint32(b[8+4:], uint32(key.Field))
	return b
}
