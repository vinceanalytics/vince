package db

import (
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/model"
	"github.com/vinceanalytics/vince/internal/ro2"
	"google.golang.org/protobuf/proto"
)

const sessionLifetime = 15 * time.Minute

func (db *Config) append(e *v1.Model, b model.Batch) error {
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
	db.cache.SetWithTTL(uint64(e.Id), e, int64(proto.Size(e)), sessionLifetime)
	return nil
}
