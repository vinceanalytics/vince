package db

import (
	"time"

	"github.com/vinceanalytics/vince/internal/batch"
	"github.com/vinceanalytics/vince/internal/models"
)

const sessionLifetime = 30 * time.Minute

func (db *Config) append(e *models.Model, b *batch.Batch) error {
	hit(e)
	if cached, ok := db.cache.Get(uint64(e.Id)); ok {
		update(cached, e)
		err := b.Add(e)
		releaseEvent(e)
		return err
	}
	newSessionEvent(e)
	err := b.Add(e)
	if err != nil {
		return err
	}
	db.cache.SetWithTTL(uint64(e.Id), e, 1, sessionLifetime)
	return nil
}
