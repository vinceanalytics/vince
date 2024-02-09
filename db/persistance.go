package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

var ErrKeyNotFound = errors.New("db: key not found")

type Storage interface {
	Set(key, value []byte, ttl time.Duration) error
	Get(key []byte, value func([]byte) error) error
	GC() error
	Close() error
}

type KV struct {
	db *badger.DB
}

type dbKey struct{}

func With(ctx context.Context, kv Storage) context.Context {
	return context.WithValue(ctx, dbKey{}, kv)
}

func Get(ctx context.Context) Storage {
	return ctx.Value(dbKey{}).(Storage)
}

func NewKV(path string) (*KV, error) {
	db, err := badger.Open(badger.DefaultOptions(filepath.Join(path, "db")).
		WithLogger(&badgerLogger{
			log: slog.Default().With(
				slog.String("component", "key-value-store"),
			),
		}))
	if err != nil {
		return nil, err
	}
	return &KV{db: db}, nil
}

var _ Storage = (*KV)(nil)

func (kv *KV) GC() error {
	return kv.db.RunValueLogGC(0.5)
}

func (kv *KV) Close() error {
	return kv.db.Close()
}

func (kv *KV) Set(key, value []byte, ttl time.Duration) error {
	println("=>", string(key))
	e := badger.NewEntry(key, value)
	if ttl > 0 {
		e = e.WithTTL(ttl)
	}
	return kv.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(e)
	})
}

func (kv *KV) Get(key []byte, value func([]byte) error) error {
	return kv.db.View(func(txn *badger.Txn) error {
		it, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrKeyNotFound
			}
			return err
		}
		return it.Value(value)
	})
}

type badgerLogger struct {
	log *slog.Logger
}

var _ badger.Logger = (*badgerLogger)(nil)

func (b *badgerLogger) Errorf(msg string, args ...interface{}) {
	b.log.Error(fmt.Sprintf(msg, args...))
}
func (b *badgerLogger) Warningf(msg string, args ...interface{}) {
	b.log.Warn(fmt.Sprintf(msg, args...))
}
func (b *badgerLogger) Infof(msg string, args ...interface{}) {
	b.log.Info(fmt.Sprintf(msg, args...))
}
func (b *badgerLogger) Debugf(msg string, args ...interface{}) {
	b.log.Debug(fmt.Sprintf(msg, args...))
}

type PrefixStore struct {
	Storage
	prefix []byte
}

func NewPrefix(store Storage, prefix string) Storage {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return &PrefixStore{
		Storage: store,
		prefix:  []byte(prefix),
	}
}

func (p *PrefixStore) Set(key, value []byte, ttl time.Duration) error {
	return p.Storage.Set(p.key(key), value, ttl)
}

func (p *PrefixStore) Get(key []byte, value func([]byte) error) error {
	return p.Storage.Get(p.key(key), value)
}

func (p *PrefixStore) key(k []byte) []byte {
	return append(p.prefix, k...)
}
