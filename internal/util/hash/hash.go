// Package hash exposes hash functions using standard maphash. We use a separate
// package to initialize a single seeed that is used throughout vince.
package hash

import "hash/maphash"

var (
	seed = maphash.MakeSeed()
)

func Bytes(value []byte) uint64 {
	return maphash.Bytes(seed, value)
}

func String(key string) uint64 {
	return maphash.String(seed, key)
}
