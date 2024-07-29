package db

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
	"github.com/gernest/len64/internal/len64"
	"github.com/gernest/len64/web/db/schema"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	db      *gorm.DB
	ts      *len64.DB
	session SessionContext
	logger  *slog.Logger
	cache   *LRU

	// we rely on cache for session processing. We need to guarantee only a single
	// writer on the cache, a buffered channel help with this.
	models chan *v1.Model
}

func Open(path string) (*Config, error) {
	ts := filepath.Join(path, "ts")
	os.MkdirAll(ts, 0755)
	ops := filepath.Join(path, "ops.db")
	db, err := open(ops)
	if err != nil {
		return nil, err
	}
	series, err := len64.Open(ts)
	if err != nil {
		conn, _ := db.DB()
		conn.Close()
		return nil, err
	}
	return &Config{
		db:     db,
		ts:     series,
		logger: slog.Default(),
		cache:  newCache(16<<10, 10*time.Minute),
		models: make(chan *v1.Model, 4<<10),
	}, nil
}

func (db *Config) Get() *gorm.DB {
	return db.db
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
	close(db.models)
	x, _ := db.db.DB()
	return errors.Join(x.Close(), db.ts.Close())
}

func open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.Logger = db.Logger.LogMode(logger.Silent)
	db.SetupJoinTable(&schema.User{}, "Sites", &schema.SiteMembership{})
	db.SetupJoinTable(&schema.Site{}, "Users", &schema.SiteMembership{})
	err = db.AutoMigrate(
		&schema.Goal{},
		&schema.Invitation{},
		&schema.SharedLink{},
		&schema.SiteMembership{},
		&schema.Site{},
		&schema.User{},
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}
