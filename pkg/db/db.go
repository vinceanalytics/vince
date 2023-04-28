// Package db provide convenient access to sqlite database. This uses context.Context
// as a store that you can pass around to database operations from it.
package db

import (
	"context"

	"gorm.io/gorm"
)

type dbKey struct{}

func Set(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey{}, db)
}

func Get(ctx context.Context) *gorm.DB {
	return ctx.Value(dbKey{}).(*gorm.DB)
}

func Exists(ctx context.Context, where func(db *gorm.DB) *gorm.DB) bool {
	return ExistsDB(Get(ctx), where)
}

func ExistsDB(db *gorm.DB, where func(db *gorm.DB) *gorm.DB) bool {
	db = where(db).Select("1").Limit(1)
	var n int
	err := db.Find(&n).Error
	return err == nil && n == 1
}

// Check performs health check on the database. This make sure we can query the
// database
func Check(ctx context.Context) bool {
	return Get(ctx).Exec("SELECT 1").Error == nil
}
