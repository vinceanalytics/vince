// Package oracle stores global truths of the whole runnning system.
package oracle

import "sync/atomic"

var (
	//Records tracks total records stored in the database. This only accounts for
	// records already in the database and excludes records in the batch ingester.
	Records atomic.Uint64
)

// Shards returns the current number of shards observed in the database. This value
// starts from 1 however we store shards starting from 0. To get correct
// shard iterations it is advised to range over returned value
//
//	for shard:=range Shards()
//	...
func Shards() (v uint64) {
	v = Records.Load()
	v = (v / (1 << 20)) + 1
	return
}
