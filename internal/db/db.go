package db

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"log/slog"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"github.com/dgraph-io/badger/v4/pb"
	"github.com/vinceanalytics/vince/internal/must"
)

type key struct{}

type raftKey struct{}

func Open(ctx context.Context, path string, logLevel ...string) (context.Context, *badger.DB) {
	if len(logLevel) == 0 || logLevel[0] != "silent" {
		slog.Info("opening main storage", "path", path)
	}
	o := badger.DefaultOptions(path).
		WithLogger(badgerLogger{}).
		WithLoggingLevel(10).
		WithCompression(options.ZSTD)

	if len(logLevel) > 0 {
		switch strings.ToLower(logLevel[0]) {
		case "debug":
			o = o.WithLoggingLevel(0)
		case "info":
			o = o.WithLoggingLevel(1)
		case "warn":
			o = o.WithLoggingLevel(2)
		case "error":
			o = o.WithLoggingLevel(3)
		}
	}
	db := must.Must(badger.Open(o))(
		"failed to open badger db:", path,
	)
	obs := &Observer{ctx: ctx, Silent: len(logLevel) > 0 && logLevel[0] == "silent"}
	go func() {
		db.Subscribe(ctx, obs.Changed, []pb.Match{
			{Prefix: []byte("vince/")},
		})
	}()

	ctx = context.WithValue(ctx, observeKey{}, obs)
	return Set(ctx, &provider{db: db}), db
}

func Set(ctx context.Context, p Provider) context.Context {
	return context.WithValue(ctx, key{}, p)
}

func Get(ctx context.Context) Provider {
	return ctx.Value(key{}).(Provider)
}

func OpenRaft(ctx context.Context, path string) (context.Context, *badger.DB) {
	path = filepath.Join(path, "db")
	slog.Info("opening raft storage", "path", path)
	o := badger.DefaultOptions(path).
		WithLogger(badgerLogger{}).
		WithLoggingLevel(10).
		WithCompression(options.ZSTD)
	db := must.Must(badger.Open(o))(
		"failed to open badger db:", path,
	)
	return context.WithValue(ctx, raftKey{}, db), db
}

func GetRaft(ctx context.Context) *badger.DB {
	return ctx.Value(raftKey{}).(*badger.DB)
}

var _ badger.Logger = (*badgerLogger)(nil)

type badgerLogger struct{}

func (badgerLogger) Errorf(format string, args ...interface{}) {
	slog.Error(fmt.Sprintf(format, args...))
}
func (badgerLogger) Warningf(format string, args ...interface{}) {
	slog.Warn(fmt.Sprintf(format, args...))
}

func (badgerLogger) Infof(format string, args ...interface{}) {
	slog.Info(fmt.Sprintf(format, args...))
}

func (b badgerLogger) Debugf(format string, args ...interface{}) {
	slog.Debug(fmt.Sprintf(format, args...))
}

type Observer struct {
	ctx      context.Context
	Silent   bool
	OnChange func(context.Context, *badger.KVList) error
}

type observeKey struct{}

func SetKeyChangeObserver(ctx context.Context, obs func(context.Context, *badger.KVList) error) {
	cb := ctx.Value(observeKey{}).(*Observer)
	cb.OnChange = obs
}

func (cb *Observer) Changed(kv *badger.KVList) error {
	if cb.OnChange != nil {
		return cb.OnChange(cb.ctx, kv)
	}
	if cb.Silent {
		return nil
	}
	for _, v := range kv.Kv {
		slog.Info("Key Change", "key", string(v.Key))
	}
	return nil
}

type Provider interface {
	With(func(db *badger.DB) error) error
	NewTransaction(update bool) Txn
	Txn(update bool, f func(txn Txn) error) error
}

type Txn interface {
	Set(key, value []byte) error
	SetTTL(key, value []byte, ttl time.Duration) error

	// Get retrieves a key and calls value with the data. When the key is not found
	// and nofFound is present then notFound is called.
	Get(key []byte, value func(val []byte) error, notFound ...func() error) error
	Has(key []byte) bool
	Delete(key []byte, notFound ...func() error) error
	Close() error

	Iter(IterOpts) Iter
}

type IterOpts struct {
	Prefix         []byte
	PrefetchValues bool
	PrefetchSize   int
	Reverse        bool
}

type Iter interface {
	Rewind()
	Valid() bool
	Next()
	Key() []byte
	Value(value func([]byte) error) error
	Close()
}

type iter struct {
	i *badger.Iterator
}

var _ Iter = (*iter)(nil)

func (i *iter) Rewind() {
	i.i.Rewind()
}

func (i *iter) Valid() bool {
	return i.i.Valid()
}
func (i *iter) Next() {
	i.i.Next()
}

func (i *iter) Key() []byte {
	return i.i.Item().Key()
}

func (i *iter) Value(value func([]byte) error) error {
	return i.i.Item().Value(value)
}

func (i *iter) Close() {
	i.i.Close()
}

type txn struct {
	x *badger.Txn
}

var _ Txn = (*txn)(nil)

func (x *txn) Set(key, value []byte) error {
	return x.x.Set(key, value)
}

func (x *txn) SetTTL(key, value []byte, ttl time.Duration) error {
	e := badger.NewEntry(key, value).WithTTL(ttl)
	return x.x.SetEntry(e)
}

func (x *txn) Get(key []byte, value func([]byte) error, notFound ...func() error) error {
	it, err := x.x.Get(key)
	if err != nil {
		if len(notFound) > 0 && errors.Is(err, badger.ErrKeyNotFound) {
			return notFound[0]()
		}
		return err
	}
	return it.Value(value)
}

func (x *txn) Has(key []byte) bool {
	_, err := x.x.Get(key)
	return err == nil
}
func (x *txn) Delete(key []byte, notFound ...func() error) error {
	err := x.x.Delete(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			if len(notFound) > 0 {
				return notFound[0]()
			}
		}
		return err
	}
	return nil
}

func (x *txn) Close() error {
	return x.x.Commit()
}

func (x *txn) Iter(o IterOpts) Iter {
	i := x.x.NewIterator(badger.IteratorOptions{
		Prefix:         o.Prefix,
		PrefetchValues: o.PrefetchValues,
		Reverse:        o.Reverse,
		PrefetchSize:   o.PrefetchSize,
	})
	return &iter{i: i}
}

type provider struct {
	db *badger.DB
}

func (p *provider) With(f func(db *badger.DB) error) error {
	return f(p.db)
}

func (p *provider) View(f func(txn *badger.Txn) error) error {
	return p.With(func(db *badger.DB) error {
		return db.View(f)
	})
}

func (p *provider) Update(f func(txn *badger.Txn) error) error {
	return p.With(func(db *badger.DB) error {
		return db.Update(f)
	})
}

func (p *provider) NewTransaction(update bool) Txn {
	return &txn{x: p.db.NewTransaction(update)}
}

func (p *provider) Txn(update bool, f func(txn Txn) error) error {
	x := p.NewTransaction(update)
	err := f(x)
	x.Close()
	return err
}
