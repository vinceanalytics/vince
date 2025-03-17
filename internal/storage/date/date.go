package date

import (
	"time"
)

const (
	minute   = 60
	hour     = minute * 60
	day      = hour * 24
	nsPerSec = 1e9
	maxUnix  = (1<<16)*day - 1
)

type Date uint16

func New(t time.Time) Date {
	s := t.Unix()
	_, offset := t.Zone()
	s += int64(offset)
	checkRange(s)
	return Date(s / day)
}

func (d Date) Time() time.Time {
	return time.Unix(int64(d)*day, 0)
}

func checkRange(seconds int64) {
	if seconds >= 0 && seconds <= maxUnix {
		return
	}
	panic("date out of range")
}
