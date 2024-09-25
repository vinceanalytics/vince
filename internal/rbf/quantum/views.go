package quantum

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Field struct {
	buffer [14]byte
	full   []byte
	prefix int
}

var fieldPool = sync.Pool{New: func() any {
	return new(Field)
}}

func NewField() *Field {
	f := fieldPool.Get().(*Field)
	return f
}

func (f *Field) Views(name string, fn func(view string)) {
	f.name(name)
	for _, unit := range lengths {
		fn(string(f.fmt(unit)))
	}
}

func (f *Field) Reset() {
	clear(f.full)
	f.full = f.full[:0]
}

func (f *Field) Release() {
	f.Reset()
	fieldPool.Put(f)
}

var lengths = []int{
	6,  //month
	8,  //week
	10, // day
	12, //hour
	14, //minute
}

func (f *Field) ViewsByTimeInto(t time.Time) {
	f.into(t)
}

func (f *Field) name(name string) {
	f.full = append(f.full[:0], []byte(name)...)
	f.full = append(f.full, '_')
	f.prefix = len(f.full)
}

func (f *Field) fmt(n int) []byte {
	f.full = append(f.full[:f.prefix], f.buffer[:n]...)
	return f.full
}

func (f *Field) Month(name string, start, end time.Time, fn func([]byte) error) error {
	return f.timeRange(name, start, end, f.month, fn)
}

func (f *Field) Week(name string, start, end time.Time, fn func([]byte) error) error {
	return f.timeRange(name, start, end, f.week, fn)
}

func (f *Field) Day(name string, start, end time.Time, fn func([]byte) error) error {
	return f.timeRange(name, start, end, f.day, fn)
}

func (f *Field) Hour(name string, start, end time.Time, fn func([]byte) error) error {
	return f.timeRange(name, start, end, f.hour, fn)
}

func (f *Field) Minute(name string, start, end time.Time, fn func([]byte) error) error {
	return f.timeRange(name, start, end, f.minute, fn)
}

func (f *Field) month(start, end time.Time, fn func(time.Time, int) error) error {
	t := beginOfMonth(start)
	end = endOfMonth(end)
	for t.Before(end) {
		err := fn(t, 6)
		if err != nil {
			return err
		}
		t = addMonth(t)
	}
	return nil
}

func (f *Field) week(start, end time.Time, fn func(time.Time, int) error) error {
	t := beginOfWeek(start)
	end = endOfWeek(end)
	for t.Before(end) {
		err := fn(t, 8)
		if err != nil {
			return err
		}
		t = t.AddDate(0, 0, 7)
	}
	return nil
}

func beginOfWeek(ts time.Time) time.Time {
	return beginOfDay(ts).AddDate(0, 0, -int(ts.Weekday()))
}

func endOfWeek(ts time.Time) time.Time {
	return beginOfWeek(ts).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

func (f *Field) day(start, end time.Time, fn func(time.Time, int) error) error {
	t := beginOfDay(start)
	end = endOfDay(end)
	for t.Before(end) {
		err := fn(t, 10)
		if err != nil {
			return err
		}
		t = t.AddDate(0, 0, 1)
	}
	return nil
}

func beginOfMonth(ts time.Time) time.Time {
	y, m, _ := ts.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, ts.Location())
}

func endOfMonth(ts time.Time) time.Time {
	return beginOfMonth(ts).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func (f *Field) hour(start, end time.Time, fn func(time.Time, int) error) error {
	t := beginOfHour(start)
	end = endOfHour(end)
	for t.Before(end) {
		err := fn(t, 12)
		if err != nil {
			return err
		}
		t = t.Add(time.Hour)
	}
	return nil
}

func beginOfDay(ts time.Time) time.Time {
	y, m, d := ts.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, ts.Location())
}

func endOfDay(ts time.Time) time.Time {
	y, m, d := ts.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), ts.Location())
}

func (f *Field) minute(start, end time.Time, fn func(time.Time, int) error) error {
	t := start.Truncate(time.Minute)
	end = end.Truncate(time.Minute).Add(time.Minute - time.Nanosecond)
	for t.Before(end) {
		err := fn(t, 14)
		if err != nil {
			return err
		}
		t = t.Add(time.Minute)
	}
	return nil
}

func beginOfHour(ts time.Time) time.Time {
	y, m, d := ts.Date()
	return time.Date(y, m, d, ts.Hour(), 0, 0, 0, ts.Location())
}

func endOfHour(ts time.Time) time.Time {
	return beginOfHour(ts).Add(time.Hour - time.Nanosecond)
}

func nextHourGTE(t time.Time, end time.Time) bool {
	next := t.Add(time.Hour)
	y1, m1, d1 := next.Date()
	y2, m2, d2 := end.Date()
	if (y1 == y2) && (m1 == m2) && (d1 == d2) {
		return true
	}
	return end.After(next)
}

func (f *Field) into(t time.Time) []byte {
	date := f.buffer[:]
	y, m, d := t.Date()
	h := t.Hour()
	mn := t.Minute()
	_, w := t.ISOWeek()
	// Did you know that Sprintf, Printf, and other things like that all
	// do allocations, and that doing allocations in a tight loop like this
	// is stunningly expensive? viewsByTime was 25% of an ingest test's
	// total CPU, not counting the garbage collector overhead. This is about
	// 3%. No, I'm not totally sure that justifies it.
	if y < 1000 {
		ys := fmt.Sprintf("%04d", y)
		copy(date[0:4], []byte(ys))
	} else if y >= 10000 {
		// This is probably a bad answer but there isn't really a
		// good answer.
		ys := fmt.Sprintf("%04d", y%1000)
		copy(date[0:4], []byte(ys))
	} else {
		strconv.AppendInt(date[:0], int64(y), 10)
	}
	//month
	date[4] = '0' + byte(m/10)
	date[5] = '0' + byte(m%10)

	//week
	date[6] = '0' + byte(w/10)
	date[7] = '0' + byte(w%10)

	//day
	date[8] = '0' + byte(d/10)
	date[9] = '0' + byte(d%10)

	//hour
	date[10] = '0' + byte(h/10)
	date[11] = '0' + byte(h%10)

	//minute
	date[12] = '0' + byte(mn/10)
	date[13] = '0' + byte(mn%10)
	return date
}

func (f *Field) timeRange(name string, start, end time.Time, gen func(start, end time.Time, fn func(time.Time, int) error) error, cb func([]byte) error) error {
	f.name(name)
	return gen(start, end, func(t time.Time, size int) error {
		f.into(t)
		return cb(f.fmt(size))
	})
}
