package quantum

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Field struct {
	buffer [12]byte
	iso    [9]byte
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
	f.full = append(f.full[:f.prefix], f.iso[:]...)
	fn(string(f.full))
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
	8,  //day
	10, // hour
	12, // minute
}

var bufferPool = &sync.Pool{New: func() any { return new(bytes.Buffer) }}

func Parse(view []byte) string {
	b := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		bufferPool.Put(b)
	}()
	switch len(view) {
	case 6:
		b.Write(view[:4])
		b.WriteByte('-')
		b.Write(view[4:])
		b.WriteString("-01")
		return b.String()
	case 8:
		b.Write(view[:4])
		b.WriteByte('-')
		b.Write(view[4:6])
		b.WriteByte('-')
		b.Write(view[6:8])
		return b.String()
	case 9:
		// iso week
		wd := view[3:]
		year, _ := strconv.Atoi(string(wd[:4]))
		week, _ := strconv.Atoi(string(wd[4:]))
		t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		offset := (t.Weekday() + 7) % 7
		t = t.Add(time.Duration(offset*24) * time.Hour)
		t = t.Add(time.Duration((week-1)*7*24) * time.Hour)
		return t.Format(time.DateOnly)
	case 10:
		b.Write(view[:4])
		b.WriteByte('-')
		b.Write(view[4:6])
		b.WriteByte('-')
		b.Write(view[6:8])
		b.WriteByte(' ')
		b.Write(view[8:10])
		b.WriteString(":00:00")
		return b.String()
	case 12:
		b.Write(view[:4])
		b.WriteByte('-')
		b.Write(view[4:6])
		b.WriteByte('-')
		b.Write(view[6:8])
		b.WriteByte(' ')
		b.Write(view[8:10])
		b.WriteByte(':')
		b.Write(view[10:12])
		b.WriteString(":00")
		return b.String()
	default:
		return ""
	}
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
	f.name(name)
	t := start
	for t.Before(end) {
		f.full = append(f.full[:f.prefix], f.intoIsoWeek(t)...)
		err := fn(f.full)
		if err != nil {
			return err
		}
		t = t.AddDate(0, 0, 7)
	}
	return nil
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
	t := start
	for t.Before(end) {
		err := fn(t, 6)
		if err != nil {
			return err
		}
		t = addMonth(t)
	}
	return nil
}

func (f *Field) day(start, end time.Time, fn func(time.Time, int) error) error {
	t := start
	for t.Before(end) {
		err := fn(t, 8)
		if err != nil {
			return err
		}
		t = t.AddDate(0, 0, 1)
	}
	return nil
}

func (f *Field) hour(start, end time.Time, fn func(time.Time, int) error) error {
	t := start
	for t.Before(end) {
		err := fn(t, 10)
		if err != nil {
			return err
		}
		t = t.Add(time.Hour)
	}
	return nil
}

func (f *Field) minute(start, end time.Time, fn func(time.Time, int) error) error {
	t := start
	for t.Before(end) {
		err := fn(t, 12)
		if err != nil {
			return err
		}
		t = t.Add(time.Minute)
	}
	return nil
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
	date[4] = '0' + byte(m/10)
	date[5] = '0' + byte(m%10)
	date[6] = '0' + byte(d/10)
	date[7] = '0' + byte(d%10)
	date[8] = '0' + byte(h/10)
	date[9] = '0' + byte(h%10)
	date[10] = '0' + byte(mn/10)
	date[11] = '0' + byte(mn%10)
	{
		// setup iso week
		f.iso[0] = 'i'
		f.iso[1] = 's'
		f.iso[2] = 'o'
		d := f.iso[3:]
		y, m := t.ISOWeek()
		if y < 1000 {
			ys := fmt.Sprintf("%04d", y)
			copy(d[0:4], []byte(ys))
		} else if y >= 10000 {
			// This is probably a bad answer but there isn't really a
			// good answer.
			ys := fmt.Sprintf("%04d", y%1000)
			copy(d[0:4], []byte(ys))
		} else {
			strconv.AppendInt(d[:0], int64(y), 10)
		}
		d[4] = '0' + byte(m/10)
		d[5] = '0' + byte(m%10)
	}
	return date
}

func (f *Field) intoIsoWeek(t time.Time) []byte {
	// setup iso week
	f.iso[0] = 'i'
	f.iso[1] = 's'
	f.iso[2] = 'o'
	d := f.iso[3:]
	y, m := t.ISOWeek()
	if y < 1000 {
		ys := fmt.Sprintf("%04d", y)
		copy(d[0:4], []byte(ys))
	} else if y >= 10000 {
		// This is probably a bad answer but there isn't really a
		// good answer.
		ys := fmt.Sprintf("%04d", y%1000)
		copy(d[0:4], []byte(ys))
	} else {
		strconv.AppendInt(d[:0], int64(y), 10)
	}
	d[4] = '0' + byte(m/10)
	d[5] = '0' + byte(m%10)
	return f.iso[:]
}

func (f *Field) timeRange(name string, start, end time.Time, gen func(start, end time.Time, fn func(time.Time, int) error) error, cb func([]byte) error) error {
	f.name(name)
	return gen(start, end, func(t time.Time, size int) error {
		f.into(t)
		return cb(f.fmt(size))
	})
}
