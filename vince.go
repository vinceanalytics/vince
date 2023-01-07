package vince

import (
	"context"

	"github.com/dgraph-io/ristretto"
	"github.com/polarsignals/frostdb"
	"github.com/sourcegraph/conc/pool"
)

type Config struct {
	DataPath string
	Cache    *ristretto.Config
}

type Vince struct {
	pool    *pool.ContextPool
	store   *frostdb.ColumnStore
	db      *frostdb.DB
	ts      *Tables
	session *SessionCache
}

func New(ctx context.Context, o *Config) (*Vince, error) {
	store, err := frostdb.New(
		frostdb.WithStoragePath(o.DataPath),
	)
	if err != nil {
		return nil, err
	}
	db, err := store.DB(context.TODO(), "vince")
	if err != nil {
		store.Close()
		return nil, err
	}
	tbl, err := NewTables(db)
	if err != nil {
		store.Close()
		db.Close()
		return nil, err
	}
	cache, err := ristretto.NewCache(o.Cache)
	if err != nil {
		store.Close()
		db.Close()
		return nil, err
	}
	v := &Vince{
		store: store,
		db:    db,
		ts:    tbl,
	}
	v.session = NewSessionCache(cache, v.ProcessSessions)
	return v, nil
}

type vinceKey struct{}

func (v *Vince) Witch(ctx context.Context) context.Context {
	return context.WithValue(ctx, vinceKey{}, v)
}

func (v *Vince) ProcessSessions(sl SessionList) {
	v.pool.Go(v.saveSession(sl))
}

func (v *Vince) saveSession(sl SessionList) func(context.Context) error {
	return func(ctx context.Context) error {
		_, err := sl.Save(ctx, v.ts)
		return err
	}
}

func From(ctx context.Context) *Vince {
	return ctx.Value(vinceKey{}).(*Vince)
}
