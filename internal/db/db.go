package db

import (
	"context"
	"fmt"
	"strings"

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
	return context.WithValue(ctx, key{}, db), db
}

func Get(ctx context.Context) Provider {
	return &provider{db: ctx.Value(key{}).(*badger.DB)}
}

func OpenRaft(ctx context.Context, path string) (context.Context, *badger.DB) {
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
	Update(f func(txn *badger.Txn) error) error
	View(f func(txn *badger.Txn) error) error
	NewTransaction(update bool) *badger.Txn
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

func (p *provider) NewTransaction(update bool) *badger.Txn {
	return p.db.NewTransaction(update)
}
