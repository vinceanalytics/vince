package db

import (
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

func (db *Config) append(e *v1.Model) {
	hit(e)
	if cached, ok := db.cache.Get(e.Id); ok {
		update(cached, e)
		db.ts.Save(e)
		return
	}
	db.ts.Save(e)
	db.cache.Add(e.Id, e)
}
