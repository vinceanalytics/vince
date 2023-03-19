package timeseries

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/gernest/vince/timex"
	"github.com/oklog/ulid/v2"
)

const (
	dateOffset    = 0
	tableOffset   = 8
	userOffset    = 9
	siteOffset    = 17
	entropyOffset = 25
)

type ID [32]byte

func (id *ID) SetTable(table byte) {
	id[tableOffset] = byte(table)
}

// Final returns id bytes without entropy. This is used as key to mike our permanent
// metrics storage.
func (id *ID) Final() []byte {
	return id[:entropyOffset]
}

func (id *ID) GetTable() byte {
	return id[tableOffset]
}

func (id *ID) SetUserID(u uint64) {
	binary.BigEndian.PutUint64(id[userOffset:], u)
}

func (id *ID) SetSiteID(u uint64) {
	binary.BigEndian.PutUint64(id[siteOffset:], u)
}

func (id *ID) GetUserID() uint64 {
	return binary.BigEndian.Uint64(id[userOffset:])
}

func (id *ID) GetSiteID() uint64 {
	return binary.BigEndian.Uint64(id[siteOffset:])
}

// SetTime converts ts to a unix date and stores it.
func (id *ID) SetTime(ts time.Time) {
	id.SetDate(timex.Date(ts))
}

func (id *ID) GetTime() time.Time {
	return time.Unix(
		int64(binary.BigEndian.Uint64(id[dateOffset:])),
		0,
	)
}

func (id *ID) SetDate(ts time.Time) {
	binary.BigEndian.PutUint64(id[dateOffset:], uint64(ts.Unix()))
}

func (id *ID) SetEntropy() {
	ulid.DefaultEntropy().Read(id[entropyOffset:])
}

// only table id ans user id
func (id *ID) Prefix() []byte {
	return id[:dateOffset]
}

func (id *ID) Release() {
	for i := 0; i < len(id); i += 1 {
		id[i] = 0
	}
	idBufPool.Put(id)
}

func (id *ID) Clone() *ID {
	x := NewID()
	copy(x[:], id[:])
	return x
}

func NewID() *ID {
	return idBufPool.Get().(*ID)
}

var idBufPool = &sync.Pool{
	New: func() any {
		var id ID
		return &id
	},
}
