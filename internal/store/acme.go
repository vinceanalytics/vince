package store

import (
	"context"

	"github.com/vinceanalytics/vince/internal/keys"
	"golang.org/x/crypto/acme/autocert"
)

var _ autocert.Cache = (*Store)(nil)

func (s *Store) Get(_ context.Context, key string) ([]byte, error) {
	tx := s.db.NewTransaction(false)
	defer tx.Discard()
	it, err := tx.Get(acmeKey(key))
	if err != nil {
		return nil, err
	}
	return it.ValueCopy(nil)
}

func (s *Store) Put(_ context.Context, key string, data []byte) error {
	tx := s.db.NewTransaction(true)
	err := tx.Set(acmeKey(key), data)
	if err != nil {
		tx.Discard()
		return err
	}
	return tx.Commit()
}

func (s *Store) Delete(_ context.Context, key string) error {
	tx := s.db.NewTransaction(true)
	err := tx.Delete(acmeKey(key))
	if err != nil {
		tx.Discard()
		return err
	}
	return tx.Commit()
}

func acmeKey(key string) []byte {
	m := make([]byte, len(keys.AcmePrefix)+len(key))
	copy(m, keys.AcmePrefix)
	copy(m[len(keys.AcmePrefix):], []byte(key))
	return m
}
