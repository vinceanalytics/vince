// package alicia defines routines to manage various keys used by  vince storage.
// We neeed specialized handling of  keys to manage the size of keys in the key value
// store becuase we never delete data and we have to advantage of having only integer
// components(exceot for specialised translation keys).
package alicia

import (
	"encoding/binary"
	"sync"
)

// Key space namespace defines prefix for different kind of keys stored
type NS byte

const (
	// Store roaring  bitap containers
	CONTAINER NS = 1 + iota

	TRANSLATE_KEY
	TRANSLATE_SEQ
	TRANSLATE_ID
	SITE
	SYSTEM
)

type Field byte

const (
	TIMESTAMP Field = 1 + iota
	ID
	BOUNCE
	SESSION
	VIEW
	DURATION
	CITY

	BROWSER
	BROWSER_VESRION
	COUNTRY
	DEVICE
	DOMAIN
	ENTRY_PAGE
	EVENT
	EXIT_PAGE
	HOST
	OS
	OS_VERSION
	PAGE
	REFERRER
	SOURCE
	UTM_CAMPAIGN
	UTM_CONTENT
	UTM_MEDIUM
	UTM_SOURCE
	UTM_TERM
	SUB1_CODE
	SUB2_CODE
)

const (
	kindOffset      = 0
	fieldOffset     = kindOffset + 1
	shardOffset     = fieldOffset + 1
	keyOffset       = shardOffset + 4
	containerOffset = keyOffset + 4
	keySize         = containerOffset + 2
)

type Key [keySize]byte

func (k *Key) NS(ns NS) *Key             { return k.b(kindOffset, byte(ns)) }
func (k *Key) Field(f uint64) *Key       { return k.b(fieldOffset, byte(f)) }
func (k *Key) Shard(shard uint64) *Key   { return k.u32(shardOffset, uint32(shard)) }
func (k *Key) Key(key uint32) *Key       { return k.u32(keyOffset, key) }
func (k *Key) Container(key uint16) *Key { return k.u16(containerOffset, key) }
func (k *Key) Bytes() []byte             { return k[:] }
func (k *Key) ShardPrefix() []byte       { return k[:keyOffset] }
func (k *Key) FieldPrefix() []byte       { return k[:shardOffset] }
func (k *Key) KeyPrefix() []byte         { return k[:containerOffset] }
func (k *Key) GetContainer() uint16      { return binary.BigEndian.Uint16(k[containerOffset:]) }

func (k *Key) b(i, v byte) *Key {
	k[i] = v
	return k
}

func (k *Key) u16(i byte, v uint16) *Key {
	binary.BigEndian.PutUint16(k[i:], v)
	return k
}

func (k *Key) u32(i byte, v uint32) *Key {
	binary.BigEndian.PutUint32(k[i:], v)
	return k
}

var keysPool = &sync.Pool{New: func() any {
	var k Key
	return &k
}}

func Get() *Key {
	return keysPool.Get().(*Key)
}

func (k *Key) Release() {
	clear(k[:])
	keysPool.Put(k)
}

func Container(b []byte) uint16 {
	return binary.BigEndian.Uint16(b[containerOffset:])
}

func Shard(b []byte) uint32 {
	return binary.BigEndian.Uint32(b[shardOffset:])
}

func (k *Key) TranslateSeq(field uint64) []byte {
	return k.NS(TRANSLATE_SEQ).Field(field).FieldPrefix()
}

func (k *Key) TranslateID(field, id uint64) []byte {
	k.NS(TRANSLATE_ID).Field(field)
	binary.BigEndian.PutUint64(k[shardOffset:], id)
	return k[:shardOffset+8]
}

func (k *Key) TranslateKey(field uint64, key []byte) []byte {
	return append(k.NS(TRANSLATE_KEY).Field(field).FieldPrefix(),
		key...)
}

func (k *Key) Site(domain string) []byte {
	k.NS(SITE)
	return append(k[:fieldOffset], []byte(domain)...)
}

func (k *Key) System() []byte {
	k.NS(SITE)
	return k[:fieldOffset]
}
