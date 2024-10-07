package db

import (
	"time"

	"github.com/vinceanalytics/vince/internal/batch"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
)

const sessionLifetime = 15 * time.Minute

func (db *Config) append(e *models.Model, b *batch.Batch) error {
	hit(e)
	if cached, ok := db.cache.Get(uint64(e.Id)); ok {
		update(cached, e)
		err := db.db.Update(func(tx *ro2.Tx) error {
			return b.Add(tx, e)
		})
		releaseEvent(e)
		return err
	}
	newSessionEvent(e)
	err := db.db.Update(func(tx *ro2.Tx) error {
		return b.Add(tx, e)
	})
	if err != nil {
		return err
	}
	db.cache.SetWithTTL(uint64(e.Id), e, 1, sessionLifetime)
	return nil
}
