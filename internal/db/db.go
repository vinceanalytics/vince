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
	"github.com/vinceanalytics/vince/internal/must"
)

// Path returns full path where badger db is created. We also store socket file
// to mysql inside dbPath directory to avoid having non badger managed files in
// its directory we create a subdirectory for badger.
func Path(dbPath string) string {
	return filepath.Join(dbPath, "vince")
}

type key struct{}

func Open(ctx context.Context, path string, logLevel ...string) (context.Context, *badger.DB) {
	path = Path(path)
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

	ctx = context.WithValue(ctx, rawKey{}, db)
	return Set(ctx, &provider{db: db}), db
}

type rawKey struct{}

func Set(ctx context.Context, p Provider) context.Context {
	return context.WithValue(ctx, key{}, p)
}

func Get(ctx context.Context) Provider {
	return ctx.Value(key{}).(Provider)
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

type Provider interface {
	NewTransaction(update bool) Transaction
}

func Update(ctx context.Context, f func(txn Transaction) error) error {
	txn := Get(ctx).NewTransaction(true)
	defer txn.Close()
	return f(txn)
}

func View(ctx context.Context, f func(txn Transaction) error) error {
	txn := Get(ctx).NewTransaction(false)
	defer txn.Close()
	return f(txn)
}

type Transaction interface {
	Set(key, value []byte, ttl time.Duration) error
	Get(key []byte, value func(val []byte) error, notFound ...func() error) error
	Delete(key []byte, notFound ...func() error) error
	Close() error
	Iter(IterOpts) Iter
}

// Has returns true if key is in the database
func Has(t Transaction, key []byte) bool {
	return t.Get(key, func(val []byte) error { return nil }) == nil
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

var _ Transaction = (*txn)(nil)

func (x *txn) Set(key, value []byte, ttl time.Duration) error {
	e := badger.NewEntry(key, value)
	if ttl != 0 {
		e = e.WithTTL(ttl)
	}
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

func (p *provider) NewTransaction(update bool) Transaction {
	return &txn{x: p.db.NewTransaction(update)}
}

func (p *provider) Txn(update bool, f func(txn Transaction) error) error {
	x := p.NewTransaction(update)
	err := f(x)
	x.Close()
	return err
}
