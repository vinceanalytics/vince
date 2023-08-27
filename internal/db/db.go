package db

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"log/slog"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"github.com/dgraph-io/badger/v4/pb"
	"github.com/vinceanalytics/vince/internal/must"
)

type key struct{}

func Open(ctx context.Context, path string, logLevel ...string) (context.Context, *badger.DB) {
	dir := filepath.Join(path, "db")
	o := badger.DefaultOptions(filepath.Join(dir, "series")).
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
		"failed to open badger db for  timeseries storage dir:", dir,
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

func Get(ctx context.Context) *badger.DB {
	return ctx.Value(key{}).(*badger.DB)
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
