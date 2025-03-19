package storage

import (
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/storage/batch"
	"github.com/vinceanalytics/vince/internal/storage/cache"
	"github.com/vinceanalytics/vince/internal/storage/translate"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

type Handle struct {
	translate *translate.Transtate
	batch     struct {
		mu       sync.RWMutex
		ba       *batch.Batch
		nxt      int64
		duration time.Duration
	}
	cache *cache.Cache
	db    *pebble.DB
}

func New(db *pebble.DB) (*Handle, error) {
	tr, err := translate.New(db)
	if err != nil {
		return nil, err
	}
	h := &Handle{translate: tr, db: db, cache: cache.New()}
	h.batch.ba = batch.New(tr)
	h.batch.duration = time.Second
	return h, err
}

func (h *Handle) Add(m *models.Model) error {
	h.cache.Update(m)

	h.batch.mu.Lock()
	defer func() {
		h.batch.mu.Unlock()
		m.Release()
	}()

	h.batch.ba.Add(m)

	if h.batch.nxt < m.Timestamp {
		// Take advantageof timestamp field to trigger flush
		err := h.unsafeFlush()
		if err != nil {
			return err
		}
		h.batch.nxt = xtime.UnixMilli(m.Timestamp).Add(h.batch.duration).Unix()
	}
	return nil
}

func (h *Handle) Close() error {
	h.batch.mu.Lock()
	defer h.batch.mu.Unlock()
	return h.unsafeFlush()
}

func (h *Handle) unsafeFlush() error {
	ba := h.db.NewBatch()
	defer ba.Close()

	err := h.batch.ba.Apply(ba)
	if err != nil {
		return err
	}

	return ba.Commit(nil)

}
