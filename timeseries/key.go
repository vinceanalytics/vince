package timeseries

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

const (
	yearOffset    = 0
	monthOffset   = 2  // 2 bytes for year
	dayOffset     = 3  // 1 byte for month
	hourOffset    = 4  // 1 byte for  day
	tableOffset   = 5  // 1 byte for hour
	metaOffset    = 6  // 1 byte for table
	userOffset    = 7  // 1 byte for metadata
	siteOffset    = 15 // 8 bytes for user id
	entropyOffset = 23 // 8 bytes for site id
	// 9 bytes for random data
)

type ID [32]byte

func (id *ID) SetTable(table byte) {
	id[tableOffset] = byte(table)
}

func (id *ID) SetMeta(table byte) {
	id[metaOffset] = byte(table)
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

func (id *ID) Hour(ts time.Time) {
	yy, mm, dd := ts.Date()
	id.setTs(yy, int(mm), dd, ts.Hour())
}

func (id *ID) Day(ts time.Time) {
	yy, mm, dd := ts.Date()
	id.setTs(yy, int(mm), dd, 0)
}

func (id *ID) Month(ts time.Time) {
	yy, mm, _ := ts.Date()
	id.setTs(yy, int(mm), 0, 0)
}

func (id *ID) Year(ts time.Time) {
	id.setTs(ts.Year(), 0, 0, 0)
}

func (id *ID) setTs(yy int, mm int, dd int, hh int) {
	binary.BigEndian.PutUint16(id[yearOffset:], uint16(yy))
	id[monthOffset] = byte(mm)
	id[dayOffset] = byte(dd)
	id[hourOffset] = byte(hh)
}

func (id *ID) SetEntropy() {
	ulid.DefaultEntropy().Read(id[entropyOffset:])
}

func (id *ID) GetTime() time.Time {
	yy := binary.BigEndian.Uint16(id[yearOffset:])
	mm := id[monthOffset]
	dd := id[dayOffset]
	hr := id[hourOffset]
	return time.Date(int(yy), time.Month(mm), int(dd), int(hr), 0, 0, 0, time.UTC)
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
