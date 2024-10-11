// Package data provides routines for interpreting byte slices to []T without a
// copy.
package data

import "unsafe"

func Bytes[T any](in []T) []byte {
	return reinterpretSlice[byte](in)
}

func Data[T any](in []byte) []T {
	return reinterpretSlice[T](in)
}

func reinterpretSlice[Out, T any](b []T) []Out {
	if cap(b) == 0 {
		return nil
	}
	out := (*Out)(unsafe.Pointer(&b[:1][0]))

	lenBytes := len(b) * int(unsafe.Sizeof(b[0]))
	capBytes := cap(b) * int(unsafe.Sizeof(b[0]))

	lenOut := lenBytes / int(unsafe.Sizeof(*out))
	capOut := capBytes / int(unsafe.Sizeof(*out))

	return unsafe.Slice(out, capOut)[:lenOut]
}
