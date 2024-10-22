package timeseries

import "hash/maphash"

var (
	seed = maphash.MakeSeed()
)

func hash(key []byte) uint64 {
	return maphash.Bytes(seed, key)
}
