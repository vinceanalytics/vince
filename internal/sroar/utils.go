/*
 * Copyright 2021 Dgraph Labs, Inc. and Contributors
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

package sroar

import (
	"log"
	"math"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
)

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

func min16(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}
func max16(a, b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}

// Returns sum of a and b. If the result overflows uint64, it returns math.MaxUint64.
func addUint64(a, b uint64) uint64 {
	if a > math.MaxUint64-b {
		return math.MaxUint64
	}
	return a + b
}

func toByteSlice(b []uint16) []byte {
	// reference: https://go101.org/article/unsafe.html
	var bs []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	hdr.Len = len(b) * 2
	hdr.Cap = hdr.Len
	hdr.Data = uintptr(unsafe.Pointer(&b[0]))
	return bs
}

// These methods (byteSliceAsUint16Slice,...) do not make copies,
// they are pointer-based (unsafe). The caller is responsible to
// ensure that the input slice does not get garbage collected, deleted
// or modified while you hold the returned slince.
////
func toUint16Slice(b []byte) (result []uint16) {
	var u16s []uint16
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&u16s))
	hdr.Len = len(b) / 2
	hdr.Cap = hdr.Len
	hdr.Data = uintptr(unsafe.Pointer(&b[0]))
	return u16s
}

// BytesToU32Slice converts the given byte slice to uint32 slice
func toUint64Slice(b []uint16) []uint64 {
	var u64s []uint64
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&u64s))
	hdr.Len = len(b) / 4
	hdr.Cap = hdr.Len
	hdr.Data = uintptr(unsafe.Pointer(&b[0]))
	return u64s
}

//go:linkname memclrNoHeapPointers runtime.memclrNoHeapPointers
func memclrNoHeapPointers(p unsafe.Pointer, n uintptr)

func Memclr(b []uint16) {
	if len(b) == 0 {
		return
	}
	p := unsafe.Pointer(&b[0])
	memclrNoHeapPointers(p, uintptr(len(b)))
}
