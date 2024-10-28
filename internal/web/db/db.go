package db

import (
	"context"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ops"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/util/lru"
)

type Config struct {
	db      *pebble.DB
	ts      *timeseries.Timeseries
	ops     *ops.Ops
	session *SessionContext
	logger  *slog.Logger
	cache   *lru.Cache[uint64, *models.Cached]
	buffer  chan *models.Model
}

func Open(db *pebble.DB, startDomains []string) (*Config, error) {

	ts := timeseries.New(db)
	ops := ops.New(db, ts, startDomains...)

	// setup session
	secret, err := ops.Web()
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Config{
		db:     db,
		ts:     ts,
		ops:    ops,
		logger: slog.Default(),
		cache:  lru.New[uint64, *models.Cached](1 << 20),
		buffer: make(chan *models.Model, 4<<10),
		session: &SessionContext{
			secret: secret,
		},
	}, nil
}

func (db *Config) PasswordMatch(pwd string) bool {
	return db.ops.VerifyPassword(pwd)
}

func (db *Config) TimeSeries() *timeseries.Timeseries {
	return db.ts
}

func (db *Config) Ops() *ops.Ops {
	return db.ops
}

func (db *Config) Pebble() *pebble.DB {
	return db.db
}

func (db *Config) Logger() *slog.Logger {
	return db.logger
}

func (db *Config) Start(ctx context.Context) {
	go db.eventsLoop(ctx)
}

func (db *Config) eventsLoop(cts context.Context) {
	db.logger.Info("starting event processing loop")
	ts := time.NewTicker(time.Second)
	defer func() {
		ts.Stop()

		err := db.ts.Save()
		if err != nil {
			db.logger.Error("applying events batch before exiting", "err", err)
		}
	}()

	for {
		select {
		case <-cts.Done():
			db.logger.Info("exiting event processing loop")
			return
		case <-ts.C:
			err := db.ts.Save()
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
	return db.ts.Close()
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
