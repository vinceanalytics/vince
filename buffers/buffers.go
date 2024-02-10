package buffers

import (
	"bytes"
	"sync"
)

type BytesBuffer struct {
	bytes.Buffer
}

func Bytes() *BytesBuffer {
	return bytesPool.Get().(*BytesBuffer)
}

func (b *BytesBuffer) Release() {
	b.Reset()
	bytesPool.Put(b)
}

var bytesPool = &sync.Pool{New: func() any { return new(BytesBuffer) }}
