package mutex

import (
	"github.com/gernest/roaring/shardwidth"
)

func Add(id uint64, value uint64) uint64 {
	return value*shardwidth.ShardWidth + (id % shardwidth.ShardWidth)
}
