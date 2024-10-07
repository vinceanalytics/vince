package ro2

import (
	"errors"

	"filippo.io/age"
	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/keys"
)

type Store struct {
	*DB
}

func Open(path string) (*Store, error) {
	return open(path)
}

func open(path string) (*Store, error) {
	db, err := newDB(path)
	if err != nil {
		return nil, err
	}
	o := &Store{
		DB: db,
	}
	return o, nil
}

func (o *Store) Web() (secret *age.X25519Identity, err error) {
	err = o.Update(func(tx *Tx) error {
		key := keys.Cookie
		it, err := tx.tx.Get(key)
		if err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
			secret, err = age.GenerateX25519Identity()
			if err != nil {
				return err
			}
			return tx.tx.Set(key, []byte(secret.String()))
		}
		return it.Value(func(val []byte) error {
			secret, err = age.ParseX25519Identity(string(val))
			return err
		})
	})
	return
}
