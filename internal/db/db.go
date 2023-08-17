package db

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"log/slog"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
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
