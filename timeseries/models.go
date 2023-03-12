package timeseries

import (
	"context"
	"path/filepath"

	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"github.com/gernest/vince/log"
)

var eventsFilterFields = []string{
	"timestamp",
	"name",
	"domain",
	"user_id",
	"session_id",
	"hostname",
	"path",
	"referrer",
	"referrer_source",
	"country_code",
	"screen_size",
	"operating_system",
	"browser",
	"utm_medium",
	"utm_source",
	"utm_campaign",
	"browser_version",
	"operating_system_version",
	"city_geo_name_id",
	"utm_content",
	"utm_term",
	"transferred_from",
	"sign",
	"is_bounce",
	"entry_page",
	"exit_page",
}

type Tables struct {
	db *badger.DB
}

func Open(ctx context.Context, allocator memory.Allocator, dir string) (*Tables, error) {
	base := filepath.Join(dir, "ts")
	o := badger.DefaultOptions(filepath.Join(base, "store")).
		WithLogger(log.Badger(ctx)).
		WithCompression(options.ZSTD)
	db, err := badger.Open(o)
	if err != nil {
		return nil, err
	}
	return &Tables{db: db}, nil
}

func (t *Tables) Close() (err error) {
	err = t.db.Close()
	return
}

type tablesKey struct{}

func Set(ctx context.Context, t *Tables) context.Context {
	return context.WithValue(ctx, tablesKey{}, t)
}

func Get(ctx context.Context) *Tables {
	return ctx.Value(tablesKey{}).(*Tables)
}
