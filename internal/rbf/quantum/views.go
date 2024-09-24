package quantum

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Field struct {
	buffer [14]byte
	ts     [5][]byte
	full   []byte
}

var fieldPool = sync.Pool{New: func() any {
	return new(Field)
}}

func NewField() *Field {
	f := fieldPool.Get().(*Field)
	return f
}

func (f *Field) Views(name string, fn func(view string)) {
	f.full = append(f.full[:0], []byte(name)...)
	f.full = append(f.full, '_')
	n := len(f.full)
	for i := range f.ts {
		f.full = append(f.full[:n], f.ts[i]...)
		fn(string(f.full))
	}
}

func (f *Field) Reset() {
	clear(f.buffer[:])
	for i := range f.ts {
		clear(f.ts[i])
	}
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
	fullBuf := f.buffer
	date := fullBuf[:]
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
	for i, unit := range lengths {
		f.ts[i] = append(f.ts[i][:0], fullBuf[:unit]...)
	}
}
