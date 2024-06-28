package db

import "time"

type Search interface {
	View(ts time.Time) View
}

type View interface {
	Shard(tx *Tx, view string, shard uint64) error
}
