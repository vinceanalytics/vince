package timeseries

import (
	"bytes"
	"time"

	"github.com/oklog/ulid/v2"
)

// Table + User + Date + Random
// 0:1 1:7 7:13 13:32
const size = 1 + 6 + 6 + 19

const (
	tableOffset   = 0
	userOffset    = 1
	dateOffset    = 7
	entropyOffset = 13
)

type TableID byte

const (
	EVENTS TableID = iota
	SESSIONS
)

// Similar to ulid.ULID. With addition of TableID and
type ID [size]byte

func (id *ID) SetTime(ts time.Time) {
	id.uint64(dateOffset, uint64(toDate(ts).Unix()))
}

func (id *ID) SetTable(table TableID) {
	(*id)[tableOffset] = byte(table)
}

func (id *ID) Entropy() {
	ulid.DefaultEntropy().Read((*id)[entropyOffset:])
}

func (id *ID) SetUserID(u uint64) {
	id.uint64(userOffset, u)
}

func (id *ID) uint64(offset int, u uint64) {
	(*id)[offset+0] = byte(u >> 40)
	(*id)[offset+1] = byte(u >> 32)
	(*id)[offset+2] = byte(u >> 24)
	(*id)[offset+3] = byte(u >> 16)
	(*id)[offset+4] = byte(u >> 8)
	(*id)[offset+5] = byte(u)
}

func (id *ID) read(offset int) uint64 {
	return uint64(id[offset+5]) | uint64(id[offset+4])<<8 |
		uint64(id[offset+3])<<16 | uint64(id[offset+2])<<24 |
		uint64(id[offset+1])<<32 | uint64(id[offset+0])<<40
}

func (id *ID) Time() time.Time {
	return time.Unix(int64(id.read(dateOffset)), 0)
}

func (id *ID) UserID() uint64 {
	return id.read(userOffset)
}

func (id *ID) Table() TableID {
	return TableID(id[tableOffset])
}

// only table id ans user id
func (id *ID) Prefix() []byte {
	return id[:dateOffset]
}

func (id *ID) PrefixWithDate() []byte {
	return id[:entropyOffset]
}

func (id *ID) Compare(other *ID) int {
	return bytes.Compare(id[:], other[:])
}
