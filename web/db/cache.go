package db

import (
	v1 "github.com/gernest/len64/gen/go/len64/v1"
)

func (db *Config) append(e *v1.Model) {
	// This is the only part accessing cache and it is guaranteed to be called in
	// only a single goroutine.
	//
	// Removing locks will yield better performance buf maintaining a single thread
	// lru cache is a big burden. This comment is a signal that below code is a low
	// hanging performance target for write heavy workload.
	hit(e)
	if cached, _ := db.cache.Get(e.Id); cached != nil {
		// Id doesn't matter if cached item has expired, we can still take advantage
		// of it and keep tracking our active session until it is purged from the
		// cache.
		update(cached, e)
		db.ts.Save(e)
		return
	}
	db.ts.Save(e)
	db.cache.Add(e.Id, e)
}
