package db

import (
	"github.com/gernest/len64/web/db/schema"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	db *gorm.DB
}

func Open(path string) (*Config, error) {
	db, err := open(path)
	if err != nil {
		return nil, err
	}
	return &Config{db: db}, nil
}

func (db *Config) Get() *gorm.DB {
	return db.db
}

func (db *Config) Close() error {
	x, _ := db.db.DB()
	return x.Close()
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

func CloseDB(db *gorm.DB) error {
	x, _ := db.DB()
	return x.Close()
}
