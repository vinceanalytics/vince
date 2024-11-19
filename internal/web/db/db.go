package db

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/geo"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ops"
	"github.com/vinceanalytics/vince/internal/shards"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/util/lru"
	"github.com/vinceanalytics/vince/internal/util/oracle"
)

type Config struct {
	lo      *location.Location
	geo     *geo.Geo
	db      *shards.DB
	ts      *timeseries.Timeseries
	ops     *ops.Ops
	session *SessionContext
	logger  *slog.Logger
	cache   *lru.Cache[uint64, *models.Cached]
	buffer  chan *models.Model
}

func Open(db *shards.DB, startDomains []string) (*Config, error) {

	lo := location.New()
	g, err := geo.New(lo, oracle.DataPath)
	if err != nil {
		return nil, err
	}
	ts := timeseries.New(db, lo)
	ops := ops.New(db.Get(), ts, startDomains...)

	// setup session
	secret, err := ops.Web()
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Config{
		lo:     lo,
		geo:    g,
		db:     db,
		ts:     ts,
		ops:    ops,
		logger: slog.Default(),
		cache:  lru.New[uint64, *models.Cached](1 << 20),
		buffer: make(chan *models.Model, 1),
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

func (db *Config) Location() *location.Location {
	return db.lo
}

func (db *Config) Ops() *ops.Ops {
	return db.ops
}

func (db *Config) Pebble() *pebble.DB {
	return db.db.Get()
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
	return errors.Join(
		db.ts.Close(),
		db.geo.Close(),
		db.lo.Close(),
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
