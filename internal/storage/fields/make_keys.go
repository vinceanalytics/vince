package fields

import (
	"encoding/binary"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

const (
	PrefixOffset    = 0
	FieldOffset     = PrefixOffset + 1
	DataTypeOffset  = FieldOffset + 1
	ShardOffset     = DataTypeOffset + 1
	ContainerOffset = ShardOffset + 8
	DataKeySize     = ContainerOffset + 8
)

const (
	TranslationShardOffset = FieldOffset + 1
	TranslationKeyOffset   = TranslationShardOffset + 8
	TranslationIDOffset    = TranslationShardOffset + 8
	TranslationIDKeySize   = TranslationIDOffset + 8
)

type Data struct {
	Shard     uint64
	Container uint64
	Field     v1.Field
	DataType  v1.DataType
}

type DataKey [DataKeySize]byte

func (d *DataKey) SetField(field v1.Field) {
	d[FieldOffset] = byte(field)
}
func (d *DataKey) SetDataType(kind v1.DataType) {
	d[DataTypeOffset] = byte(kind)
}

func (d *DataKey) SetShard(shard uint64) {
	binary.BigEndian.PutUint64(d[ShardOffset:], shard)
}

func (d *DataKey) SetContainer(key uint64) {
	binary.BigEndian.PutUint64(d[ContainerOffset:], key)
}

func (d *DataKey) Make(field v1.Field, kind v1.DataType, shard, key uint64) {
	d[FieldOffset] = byte(field)
	d[DataTypeOffset] = byte(kind)
	binary.BigEndian.PutUint64(d[ShardOffset:], shard)
	binary.BigEndian.PutUint64(d[ContainerOffset:], key)
}

func MakeTranslationKey(field v1.Field, shard uint64, value []byte) []byte {
	o := make([]byte, TranslationKeyOffset+len(value))
	o[PrefixOffset] = byte(v1.Prefix_TranslateKey)
	o[FieldOffset] = byte(field)
	binary.BigEndian.PutUint64(o[TranslationShardOffset:], shard)
	copy(o[TranslationKeyOffset:], value)
	return o
}

func MakeTranslationID(field v1.Field, shard uint64, id uint64) []byte {
	o := make([]byte, TranslationIDKeySize)
	o[PrefixOffset] = byte(v1.Prefix_TranslateID)
	o[FieldOffset] = byte(field)
	binary.BigEndian.PutUint64(o[TranslationShardOffset:], shard)
	binary.BigEndian.PutUint64(o[TranslationKeyOffset:], id)
	return o
}

func MakeSeqKey() []byte {
	return []byte{byte(v1.Prefix_SEQ)}
}
