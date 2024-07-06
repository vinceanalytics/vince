package db

import (
	"path/filepath"

	"github.com/gernest/rbf/dsl"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

type DB struct {
	db *dsl.Store[*v1.Data]
}

func New(path string) (*DB, error) {
	base := filepath.Join(path, "v1alpha1")
	db, err := dsl.New[*v1.Data](base)
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Append(data []*v1.Data) error {
	return db.db.Append(data)
}
