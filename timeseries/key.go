package timeseries

import (
	"encoding/binary"
	"sync"
	"time"
)

const (
	yearOffset  = 0
	monthOffset = 2  // 2 bytes for year
	dayOffset   = 3  // 1 byte for month
	hourOffset  = 4  // 1 byte for  day
	tableOffset = 5  // 1 byte for hour
	metaOffset  = 6  // 1 byte for table
	userOffset  = 7  // 1 byte for metadata
	siteOffset  = 15 // 8 bytes for user id
)

type ID [24]byte

func (id *ID) SetTable(table byte) *ID {
	id[tableOffset] = byte(table)
	return id
}

func (id *ID) SetMeta(table byte) *ID {
	id[metaOffset] = byte(table)
	return id
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

func (id *ID) Hour(ts time.Time) *ID {
	yy, mm, dd := ts.Date()
	return id.setTs(yy, int(mm), dd, ts.Hour())
}

func (id *ID) Day(ts time.Time) *ID {
	yy, mm, dd := ts.Date()
	return id.setTs(yy, int(mm), dd, 0)
}

func (id *ID) Month(ts time.Time) *ID {
	yy, mm, _ := ts.Date()
	return id.setTs(yy, int(mm), 0, 0)
}

func (id *ID) Year(ts time.Time) *ID {
	return id.setTs(ts.Year(), 0, 0, 0)
}

func (id *ID) setTs(yy int, mm int, dd int, hh int) *ID {
	binary.BigEndian.PutUint16(id[yearOffset:], uint16(yy))
	id[monthOffset] = byte(mm)
	id[dayOffset] = byte(dd)
	id[hourOffset] = byte(hh)
	return id
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
