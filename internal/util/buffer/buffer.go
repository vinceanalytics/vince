/*
 * Copyright 2020 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package buffer

import (
	"encoding/binary"
	"fmt"
	"log"
	"slices"
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	defaultCapacity = 64
	defaultTag      = "buffer"
)

type Source interface {
	Data() []byte
	Grow(size int) []byte
	Release() error
}

type Slice struct {
	data []byte
}

func (s *Slice) Data() []byte   { return s.data }
func (s *Slice) Release() error { return nil }
func (s *Slice) Grow(size int) []byte {
	s.data = slices.Grow(s.data, size)[:size]
	return s.data
}

// Buffer is equivalent of bytes.Buffer without the ability to read. It is NOT thread-safe.
//
// MaxSize can be set to limit the memory usage.
type Buffer struct {
	source  Source
	padding uint64 // number of starting bytes used for padding
	offset  uint64 // used length of the buffer
	buf     []byte // backing slice for the buffer
	curSz   int    // capacity of the buffer
	maxSz   int    // causes a panic if the buffer grows beyond this size
}

func NewBuffer(source Source, capacity int) *Buffer {

	return &Buffer{
		source: source,
		buf:    source.Data(),
		curSz:  max(capacity, defaultCapacity),
	}
}

func (b *Buffer) WithMaxSize(size int) *Buffer {
	b.maxSz = size
	return b
}

func (b *Buffer) IsEmpty() bool {
	return int(b.offset) == b.StartOffset()
}

// LenWithPadding would return the number of bytes written to the buffer so far
// plus the padding at the start of the buffer.
func (b *Buffer) LenWithPadding() int {
	return int(atomic.LoadUint64(&b.offset))
}

// LenNoPadding would return the number of bytes written to the buffer so far
// (without the padding).
func (b *Buffer) LenNoPadding() int {
	return int(atomic.LoadUint64(&b.offset) - b.padding)
}

// Bytes would return all the written bytes as a slice.
func (b *Buffer) Bytes() []byte {
	off := atomic.LoadUint64(&b.offset)
	return b.buf[b.padding:off]
}

// Grow would grow the buffer to have at least n more bytes. In case the buffer is at capacity, it
// would reallocate twice the size of current capacity + n, to ensure n bytes can be written to the
// buffer without further allocation. In UseMmap mode, this might result in underlying file
// expansion.
func (b *Buffer) Grow(n int) {
	if b.buf == nil {
		panic("z.Buffer needs to be initialized before using")
	}
	if b.maxSz > 0 && int(b.offset)+n > b.maxSz {
		err := fmt.Errorf(
			"z.Buffer max size exceeded: %d offset: %d grow: %d", b.maxSz, b.offset, n)
		panic(err)
	}
	if int(b.offset)+n < b.curSz {
		return
	}

	// Calculate new capacity.
	growBy := b.curSz + n
	// Don't allocate more than 1GB at a time.
	if growBy > 1<<30 {
		growBy = 1 << 30
	}
	// Allocate at least n, even if it exceeds the 1GB limit above.
	if n > growBy {
		growBy = n
	}
	b.curSz += growBy
	b.buf = b.source.Grow(b.curSz)
}

// Allocate is a way to get a slice of size n back from the buffer. This slice can be directly
// written to. Warning: Allocate is not thread-safe. The byte slice returned MUST be used before
// further calls to Buffer.
func (b *Buffer) Allocate(n int) []byte {
	b.Grow(n)
	off := b.offset
	b.offset += uint64(n)
	return b.buf[off:int(b.offset)]
}

// AllocateOffset works the same way as allocate, but instead of returning a byte slice, it returns
// the offset of the allocation.
func (b *Buffer) AllocateOffset(n int) int {
	b.Grow(n)
	b.offset += uint64(n)
	return int(b.offset) - n
}

func (b *Buffer) writeLen(sz int) {
	buf := b.Allocate(8)
	binary.BigEndian.PutUint64(buf, uint64(sz))
}

// SliceAllocate would encode the size provided into the buffer, followed by a call to Allocate,
// hence returning the slice of size sz. This can be used to allocate a lot of small buffers into
// this big buffer.
// Note that SliceAllocate should NOT be mixed with normal calls to Write.
func (b *Buffer) SliceAllocate(sz int) []byte {
	b.Grow(8 + sz)
	b.writeLen(sz)
	return b.Allocate(sz)
}

func (b *Buffer) StartOffset() int {
	return int(b.padding)
}

func (b *Buffer) WriteSlice(slice []byte) {
	dst := b.SliceAllocate(len(slice))
	assert(len(slice) == copy(dst, slice))
}

func assert(b bool) {
	if !b {
		log.Fatalf("%+v", errors.Errorf("Assertion failure"))
	}
}
func check(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
func check2(_ interface{}, err error) {
	check(err)
}

func (b *Buffer) Data(offset int) []byte {
	if offset > b.curSz {
		panic("offset beyond current size")
	}
	return b.buf[offset:b.curSz]
}

// Write would write p bytes to the buffer.
func (b *Buffer) Write(p []byte) (n int, err error) {
	n = len(p)
	b.Grow(n)
	assert(n == copy(b.buf[b.offset:], p))
	b.offset += uint64(n)
	return n, nil
}

// Reset would reset the buffer to be reused.
func (b *Buffer) Reset() {
	b.offset = uint64(b.StartOffset())
}

// Release would free up the memory allocated by the buffer. Once the usage of buffer is done, it is
// important to call Release, otherwise a memory leak can happen.
func (b *Buffer) Release() error {
	return b.source.Release()
}
