package db

import (
	"github.com/vinceanalytics/vince/internal/models"
)

func (db *Config) append(e *models.Model) error {
	hit(e)
	if cached, ok := db.cache.Get(e.Id); ok {
		if m := e.Update(cached); m != nil {
			db.cache.Set(e.Id, m)
		}
		err := db.ts.Add(e)
		releaseEvent(e)
		return err
	}
	newSessionEvent(e)
	err := db.ts.Add(e)
	if err != nil {
		return err
	}
	db.cache.Set(e.Id, e.Cached())
	return nil
}
