package fields

import (
	"encoding/binary"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

const (
	FieldOffset     = 0
	DataTypeOffset  = FieldOffset + 1
	ShardOffset     = DataTypeOffset + 1
	ContainerOffset = ShardOffset + 8
	DataKeySize     = ContainerOffset + 8
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
