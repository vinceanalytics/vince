package len64

import (
	"github.com/cockroachdb/pebble"
)

type DB struct {
	db *pebble.DB
}
