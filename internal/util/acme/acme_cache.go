package acme

import (
	"bytes"
	"context"
	"errors"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/util/data"
	"golang.org/x/crypto/acme/autocert"
)

type Cache struct {
	db *pebble.DB
}

func New(db *pebble.DB) *Cache {
	return &Cache{db: db}
}

var _ autocert.Cache = (*Cache)(nil)

func (a *Cache) Get(_ context.Context, key string) (rs []byte, err error) {
	err = data.Get(a.db, encoding.ACME([]byte(key)), func(val []byte) error {
		rs = bytes.Clone(val)
		return nil
	})
	if errors.Is(err, pebble.ErrNotFound) {
		return nil, autocert.ErrCacheMiss
	}
	return
}

func (a *Cache) Put(_ context.Context, key string, data []byte) error {
	return a.db.Set(encoding.ACME([]byte(key)), data, nil)
}

func (a *Cache) Delete(_ context.Context, key string) error {
	return a.db.Delete(encoding.ACME([]byte(key)), nil)
}
