package roaring

import (
	"runtime"
	"sync"
)

func (ra *Bitmap) Batch() Batch {
	N := ra.keys.numKeys()
	P := runtime.NumCPU()

	width := N / P
	rem := N % P
	cnt := 0

	batch := make([][]uint64, P)
	for i := 0; i < P; i++ {
		batch[i] = ra.keys[indexNodeStart : indexNodeStart+N*2]
		n := width
		if i < rem {
			n = width + 1
		}
		batch[i] = batch[i][cnt : cnt+2*n]
		cnt = cnt + 2*n
	}
	for i := range P {
		if len(batch[i]) == 0 {
			batch = batch[:i]
			break
		}
	}
	return batch
}

type Batch [][]uint64

type Next func() (uint64, bool)
type BatchAction func(idx int, next Next)

func (b Batch) Run(ra *Bitmap, action BatchAction) {
	if len(b) == 0 {
		return
	}
	if len(b) == 1 {
		if len(b[0]) == 2 {
			// single container. Divide work  based on the cardinality
			key := b[0][0]
			offset := b[0][1]
			c := ra.getContainer(offset)

			var all []uint16
			switch c[indexType] {
			case typeArray:
				a := array(c)
				all = a.all()
			case typeBitmap:
				b := bitmap(c)
				all = b.all()
			default:
			}
			if len(all) > 0 {
				var wg sync.WaitGroup
				N := len(all)
				P := runtime.NumCPU()

				x := N / P

				remainder := N - (x * P)
				pos := 0
				for i := range P {
					off := pos + x
					if i == P-1 {
						off += remainder
					}
					start := pos
					pos = off
					i := i
					nxt := nextSlice(all, key, start, off)
					wg.Add(1)
					go work(i, action, nxt, &wg)
				}
				wg.Wait()
			}

			return
		}
		// for a single key, we ignore the number of CPU and process each
		// container in its own goroutine
		var wg sync.WaitGroup
		for i := 0; i < len(b[0]); i++ {
			key := b[0][i]
			offset := b[0][i+1]
			i++
			nxt := next(ra, key, offset)
			wg.Add(1)
			i := i
			go work(i, action, nxt, &wg)
		}
		wg.Wait()
		return
	}
	var wg sync.WaitGroup
	for i := range b {
		nxt := nextRange(ra, b[i])
		wg.Add(1)
		i := i
		go work(i, action, nxt, &wg)
	}
	wg.Wait()
}

func nextSlice(ra []uint16, key uint64, start, end int) Next {
	return func() (uint64, bool) {
		if start < end {
			o := key | uint64(ra[start])
			start++
			return o, true
		}
		return 0, false
	}
}

func nextRange(ra *Bitmap, keys []uint64) Next {
	pos := 0
	nxt := next(ra, keys[pos], keys[pos+1])
	pos += 2
	return func() (uint64, bool) {
		if o, ok := nxt(); ok {
			return o, ok
		}
		if pos < len(keys) {
			nxt = next(ra, keys[pos], keys[pos+1])
			pos += 2
			return nxt()
		}
		return 0, false
	}
}

func work(idx int, action BatchAction, next Next, wg *sync.WaitGroup) {
	action(idx, next)
	wg.Done()
}

func next(ra *Bitmap, key, offset uint64) func() (uint64, bool) {
	c := ra.getContainer(offset)

	var all []uint16
	switch c[indexType] {
	case typeArray:
		a := array(c)
		all = a.all()
	case typeBitmap:
		b := bitmap(c)
		all = b.all()
	default:
		return func() (uint64, bool) { return 0, false }
	}
	pos := 0
	size := len(all)
	return func() (uint64, bool) {
		if pos < size {
			o := key | uint64(all[pos])
			pos++
			return o, true
		}
		return 0, false
	}
}
