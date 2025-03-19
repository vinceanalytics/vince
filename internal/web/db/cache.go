package db

import (
	"github.com/vinceanalytics/vince/internal/models"
)

func (db *Config) append(e *models.Model) error {
	e.Hit()
	if cached, ok := db.cache.Get(e.Id); ok {
		if m := e.Update(cached); m != nil {
			db.cache.Add(e.Id, m)
		}
		err := db.ts.Add(e)
		e.Release()
		return err
	}
	e.NewSession()
	err := db.ts.Add(e)
	if err != nil {
		return err
	}
	db.cache.Add(e.Id, e.Cached())
	return nil
}
