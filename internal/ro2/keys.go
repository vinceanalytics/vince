package ro2

import (
	"encoding/binary"
	"fmt"
	"sync"
)

const (
	kindOffset      = 0
	fieldOffset     = kindOffset + 2
	shardOffset     = fieldOffset + 8
	keyOffset       = shardOffset + 8
	containerOffset = keyOffset + 4
	keySize         = containerOffset + 2
)

type Kind byte

const (
	ROAR byte = iota
	TRANSLATE_KEY
	TRANSLATE_SEQ
	TRANSLATE_ID
	USER_ID
	User_EMAIL
	SITE_DOMAIN
	SYSTEM
)

type Key [keySize]byte

func (k *Key) Field() uint64 {
	return binary.BigEndian.Uint64(k[fieldOffset:])
}

func (k *Key) SetField(field uint64) *Key {
	binary.BigEndian.PutUint64(k[fieldOffset:], field)
	return k
}

func (k *Key) Shard() uint64 {
	return binary.BigEndian.Uint64(k[shardOffset:])
}

func (k *Key) ShardPrefix() []byte {
	return k[:keyOffset]
}

func (k *Key) SetShard(v uint64) *Key {
	binary.BigEndian.PutUint64(k[shardOffset:], v)
	return k
}

func (k *Key) FieldPrefix() []byte {
	return k[:shardOffset]
}

func (k *Key) Kind() Kind {
	return Kind(k[kindOffset])
}

func (k *Key) SetKind(kind Kind) *Key {
	k[kindOffset] = byte(kind)
	return k
}

func (k *Key) Key() uint32 {
	return binary.BigEndian.Uint32(k[keyOffset:])
}

func ReadKey(k []byte) (container uint16) {
	return binary.BigEndian.Uint16(k[containerOffset:])
}

func (k *Key) SetKey(v uint32) *Key {
	binary.BigEndian.PutUint32(k[keyOffset:], v)
	return k
}

func (k *Key) KeyPrefix() []byte {
	return k[:containerOffset]
}

func (k *Key) Container() uint16 {
	return binary.BigEndian.Uint16(k[containerOffset:])
}

func (k *Key) SetContainer(v uint16) *Key {
	binary.BigEndian.PutUint16(k[containerOffset:], v)
	return k
}

func (k *Key) String() string {
	return fmt.Sprintf("%d/%d/%d/%d",
		k.Shard(), k.Field(), k.Key(), k.Container())
}

var keyPool = &sync.Pool{New: func() any {
	var k Key
	return &k
}}

// Keys safely retains key during a write transaction. We need to ensure keys
// are valid until transactions are committed.
type Keys struct {
	keys []*Key
}

func (k *Keys) Get() *Key {
	g := keyPool.Get().(*Key)
	k.keys = append(k.keys, g)
	return g
}

func (k *Keys) Release() {
	for _, g := range k.keys {
		clear(g[:])
		keyPool.Put(g)
	}
	clear(k.keys)
	k.keys = k.keys[:0]
}
