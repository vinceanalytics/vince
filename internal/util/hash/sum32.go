package hash

import (
	"hash"
	"hash/crc32"
	"sync"
)

var pool = &sync.Pool{New: func() any { return crc32.NewIEEE() }}

func Sum32(data []byte) uint32 {
	h := pool.Get().(hash.Hash32)
	h.Write(data)
	sum := h.Sum32()
	h.Reset()
	pool.Put(h)
	return sum
}
