package storage

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/pebble"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/storage/batch"
	"github.com/vinceanalytics/vince/internal/storage/cache"
	"github.com/vinceanalytics/vince/internal/storage/translate/mapping"
)

type Handle struct {
	mapping *mapping.Mapping
	cache   *cache.Cache
	db      *pebble.DB
}

func New(db *pebble.DB, path string) (*Handle, error) {
	translatePath := filepath.Join(path, "translation")
	_ = os.MkdirAll(translatePath, 0755)
	tr, err := mapping.New(translatePath)
	if err != nil {
		return nil, err
	}
	return &Handle{mapping: tr, db: db, cache: cache.New()}, nil
}

func (h *Handle) Add(m ...*models.Model) error {
	for i := range m {
		h.cache.Update(m[i])
	}

	ba := batch.New(h.mapping, h.db.NewBatch())
	defer ba.Close()

	for i := range m {
		err := ba.Add(m[i])
		if err != nil {
			return err
		}
	}

	return ba.Commit()
}

func (b *Handle) Translate(field v1.Field, value []byte) uint64 {
	return b.mapping.Get(field, value)
}

func (b *Handle) Shards() uint64 {
	return (b.mapping.Load() / shardwidth.ShardWidth) + 1
}
