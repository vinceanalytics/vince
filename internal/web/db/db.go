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

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/lru"
	"github.com/vinceanalytics/vince/internal/ro2"
)

type Config struct {
	db      *ro2.Store
	session *SessionContext
	logger  *slog.Logger
	cache   *lru.LRU[*v1.Model]
}

func Open(path string) (*Config, error) {
	ts := filepath.Join(path, "db")
	os.MkdirAll(ts, 0755)
	ops, err := ro2.Open(ts)
	if err != nil {
		return nil, err
	}

	// setup session
	secret, err := newSession(path)
	if err != nil {
		ops.Close()
		return nil, err
	}
	return &Config{
		db:     ops,
		logger: slog.Default(),
		cache:  lru.New[*v1.Model](16 << 10),
		session: &SessionContext{
			secret: secret,
		},
	}, nil
}

func (db *Config) Get() *ro2.Store {
	return db.db
}

func (db *Config) Logger() *slog.Logger {
	return db.logger
}

func (db *Config) Start(ctx context.Context) {
	go db.db.Start(ctx)
}

func (db *Config) Close() error {
	return errors.Join(
		db.db.Close(),
	)
}

func (db *Config) HTML(w http.ResponseWriter, t *template.Template, data map[string]any) {
	db.HTMLCode(http.StatusOK, w, t, data)
}

func (db *Config) HTMLCode(code int, w http.ResponseWriter, t *template.Template, data map[string]any) {
	if data == nil {
		data = make(map[string]any)
	}
	w.Header().Set("content-type", "text/html")
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
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		db.logger.Error("rendering template", "err", err)
	}
}
