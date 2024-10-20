package db

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/dgraph-io/ristretto"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/domains"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/store"
)

type Config struct {
	config  *v1.Config
	db      *store.Store
	session *SessionContext
	logger  *slog.Logger
	cache   *ristretto.Cache[uint64, *models.Model]
	buffer  chan *models.Model
}

func Open(config *v1.Config) (*Config, error) {
	if config.DataPath != "" {
		os.MkdirAll(config.DataPath, 0755)
	}
	ops, err := store.Open(config.DataPath)
	if err != nil {
		return nil, err
	}
	cache, err := ristretto.NewCache(&ristretto.Config[uint64, *models.Model]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 20, // 1 million active sessions.
		BufferItems: 64,      // number of keys per Get buffer.
		OnEvict: func(item *ristretto.Item[*models.Model]) {
			releaseEvent(item.Value)
			item.Value = nil
		},
		OnReject: func(item *ristretto.Item[*models.Model]) {
			releaseEvent(item.Value)
			item.Value = nil
		},
	})
	if err != nil {
		return nil, err
	}
	// setup session
	secret, err := ops.Web()
	if err != nil {
		ops.Close()
		return nil, err
	}
	return &Config{
		config: config,
		db:     ops,
		logger: slog.Default(),
		cache:  cache,
		buffer: make(chan *models.Model, 4<<10),
		session: &SessionContext{
			secret: secret,
		},
	}, nil
}

func (db *Config) PasswordMatch(email, pwd string) bool {
	if db.config.Admin.Email != email {
		return *False
	}
	return subtle.ConstantTimeCompare(
		[]byte(db.config.GetAdmin().GetPassword()),
		[]byte(pwd),
	) == 1
}

func (db *Config) Get() *store.Store {
	return db.db
}

func (db *Config) GetConfig() *v1.Config {
	return db.config
}

func (db *Config) Logger() *slog.Logger {
	return db.logger
}

func (db *Config) Start(ctx context.Context) {
	domains.Reload(db.Get().Domains)
	db.Get().SetupDomains(db.config.Domains)
	if len(db.config.Domains) > 0 {
		domains.Reload(db.Get().Domains)
	}
	go db.db.Start(ctx)
	go db.eventsLoop(ctx)
}

func (db *Config) eventsLoop(cts context.Context) {
	db.logger.Info("starting event processing loop")
	ts := time.NewTicker(time.Second)
	defer ts.Stop()
	for {
		select {
		case <-cts.Done():
			db.logger.Info("exiting event processing loop")
			return
		case <-ts.C:
			err := db.db.SaveBatch()
			if err != nil {
				db.logger.Error("applying events batch", "err", err)
			}
		case e := <-db.buffer:
			err := db.append(e)
			if err != nil {
				db.logger.Error("appening event", "err", err)
			}
		}
	}
}

func (db *Config) Close() error {
	db.cache.Close()
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
