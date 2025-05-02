package shard

import (
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/gernest/roaring"
	"github.com/vinceanalytics/vince/internal/storage/translate/mapping"
)

type Shard struct {
	View    time.Time
	It      *pebble.Iterator
	Shard   uint64
	Mapping *mapping.Mapping
}

type Match struct {
	Ra *roaring.Bitmap
}
