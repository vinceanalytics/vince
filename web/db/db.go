package db

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
	"github.com/gernest/len64/internal/domains"
	"github.com/gernest/len64/internal/kv"
	"github.com/gernest/len64/internal/len64"
)

type Config struct {
	domains *domains.Cache
	db      *kv.Pebble
	ts      *len64.DB
	session SessionContext
	logger  *slog.Logger
	cache   *cache

	// we rely on cache for session processing. We need to guarantee only a single
	// writer on the cache, a buffered channel help with this.
	models chan *v1.Model
}

func Open(path string) (*Config, error) {
	ts := filepath.Join(path, "ts")
	os.MkdirAll(ts, 0755)
	series, err := len64.Open(ts)
	if err != nil {
		return nil, err
	}
	kv := kv.New(series.KV())
	doms, err := domains.New(kv)
	if err != nil {
		series.Close()
		return nil, err
	}
	return &Config{
		domains: doms,
		db:      kv,
		ts:      series,
		logger:  slog.Default(),
		cache:   newCache(16 << 10),
		models:  make(chan *v1.Model, 4<<10),
	}, nil
}

func (db *Config) Get() *kv.Pebble {
	return db.db
}

func (db *Config) Logger() *slog.Logger {
	return db.logger
}

func (db *Config) Start(ctx context.Context) error {
	go db.processEvents()
	return db.ts.Start(ctx)
}

func (db *Config) processEvents() {
	db.logger.Info("start event processing loop")
	for m := range db.models {
		db.append(m)
	}
	db.logger.Info("stopped events processing loop")
}

func (db *Config) Close() error {
	return db.ts.Close()
}

func (db *Config) HTML(w http.ResponseWriter, t *template.Template, data map[string]any) {
	db.HTMLCode(http.StatusOK, w, t, data)
}

func (db *Config) HTMLCode(code int, w http.ResponseWriter, t *template.Template, data map[string]any) {
	if data == nil {
		data = make(map[string]any)
	}
	w.WriteHeader(code)
	err := t.Execute(w, db.Context(data))
	if err != nil {
		db.logger.Error("rendering template", "err", err)
	}
}
