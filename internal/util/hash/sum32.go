package hash

import (
	"github.com/dgraph-io/badger/v4/y"
)

func Sum32(data []byte) uint32 {
	return y.Hash(data)
}
