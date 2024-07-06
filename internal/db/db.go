package db

import (
	"path/filepath"

	"github.com/gernest/rbf/dsl"
	v2 "github.com/vinceanalytics/vince/gen/go/vince/v2"
)

type DB struct {
	db *dsl.Store[*v2.Data]
}

func New(path string) (*DB, error) {
	base := filepath.Join(path, "v1alpha1")
	db, err := dsl.New[*v2.Data](base)
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Append(data []*v2.Data) error {
	return db.db.Append(data)
}
