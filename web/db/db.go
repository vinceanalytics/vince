package db

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
	"github.com/gernest/len64/internal/domains"
	"github.com/gernest/len64/internal/oracle"
	"go.etcd.io/bbolt"
)

type Config struct {
	domains *domains.Cache
	db      *bbolt.DB
	ts      *oracle.Oracle
	session SessionContext
	logger  *slog.Logger
	cache   *cache

	// we rely on cache for session processing. We need to guarantee only a single
	// writer on the cache, a buffered channel help with this.
	models chan *v1.Model

	tx *bbolt.Tx
}

func Open(path string) (*Config, error) {
	ts := filepath.Join(path, "ts")
	os.MkdirAll(ts, 0755)
	ops, err := bbolt.Open(filepath.Join(path, "main.db"), 0600, nil)
	if err != nil {
		return nil, err
	}
	db, err := oracle.New(ts)
	if err != nil {
		ops.Close()
		return nil, err
	}
	doms, err := domains.New(ops)
	if err != nil {
		ops.Close()
		db.Close()
		return nil, err
	}
	return &Config{
		domains: doms,
		db:      ops,
		ts:      db,
		logger:  slog.Default(),
		cache:   newCache(16 << 10),
		models:  make(chan *v1.Model, 4<<10),
	}, nil
}

func (db *Config) Get() *bbolt.Tx {
	return db.tx
}

func (db *Config) Logger() *slog.Logger {
	return db.logger
}

func (db *Config) Oracle() *oracle.Oracle {
	return db.ts
}

func (db *Config) Start(ctx context.Context) {
	go db.processEvents()
	db.ts.Start(ctx)
}

func (db *Config) processEvents() {
	db.logger.Info("start event processing loop")
	for m := range db.models {
		db.append(m)
	}
	db.logger.Info("stopped events processing loop")
}

func (db *Config) Close() error {
	return errors.Join(
		db.db.Close(),
		db.ts.Close(),
	)
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

func (db *Config) JSON(w http.ResponseWriter, data any) {
	db.JSONCode(http.StatusOK, w, data)
}

func (db *Config) JSONCode(code int, w http.ResponseWriter, data any) {
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		db.logger.Error("rendering template", "err", err)
	}
}
