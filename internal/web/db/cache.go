package db

import (
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

func (db *Config) append(e *v1.Model) error {
	hit(e)
	if cached, ok := db.cache.Get(uint64(e.Id)); ok {
		update(cached, e)
		return db.db.One(e)
	}
	err := db.db.One(e)
	if err != nil {
		return err
	}
	db.cache.Add(uint64(e.Id), e)
	return nil
}
