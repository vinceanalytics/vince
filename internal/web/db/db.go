package db

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/domains"
	"github.com/vinceanalytics/vince/internal/lru"
	"github.com/vinceanalytics/vince/internal/ro2"
)

var (
	buffer = flag.Duration("db.buffer", time.Minute, "Duration events are buffered before saving to disc")
)

type Config struct {
	domains *domains.Cache
	db      *ro2.Store
	session *SessionContext
	logger  *slog.Logger
	cache   *lru.LRU[*v1.Model]

	// we rely on cache for session processing. We need to guarantee only a single
	// writer on the cache, a buffered channel help with this.
	models chan *v1.Model

	disableRegistration bool
}

func Open(path string) (*Config, error) {
	ts := filepath.Join(path, "db")
	os.MkdirAll(ts, 0755)
	ops, err := ro2.Open(ts)
	if err != nil {
		return nil, err
	}
	doms := domains.New()
	ops.Domains(doms.Load())

	// setup session
	secret, err := newSession(path)
	if err != nil {
		ops.Close()
		return nil, err
	}
	return &Config{
		domains: doms,
		db:      ops,
		logger:  slog.Default(),
		cache:   lru.New[*v1.Model](16 << 10),
		models:  make(chan *v1.Model, 4<<10),
		session: &SessionContext{
			secret: secret,
		},
	}, nil
}

func (db *Config) DisableRegistration(disable bool) {
	db.disableRegistration = disable
}

func (db *Config) Get() *ro2.Store {
	return db.db
}

func (db *Config) Logger() *slog.Logger {
	return db.logger
}

func (db *Config) Start(ctx context.Context) {
	go db.processEvents(ctx)
	go db.db.Start(ctx)
}

func (db *Config) processEvents(ctx context.Context) {
	db.logger.Info("start event processing loop", "buffer", buffer.String())
	tick := time.NewTicker(*buffer)
	defer tick.Stop()
	defer db.logger.Info("stopped events processing loop")
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-db.models:
			db.append(m)
		case <-tick.C:
			err := db.db.Flush()
			if err != nil {
				slog.Error("flushing buffered events", "err", err)
			}
		}
	}
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
