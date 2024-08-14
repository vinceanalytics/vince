package ro

func MutexPosition(id, value uint64) uint64 {
	return (value*ShardWidth + (id % ShardWidth))
}
