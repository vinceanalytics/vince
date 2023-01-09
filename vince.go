package vince

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/dgraph-io/ristretto"
	"github.com/polarsignals/frostdb"
	"github.com/sourcegraph/conc/pool"
	"gorm.io/gorm"
)

const MAX_BUFFER_SIZE = 4098

type Config struct {
	DataPath string
}

type Vince struct {
	pool    *pool.ContextPool
	store   *frostdb.ColumnStore
	db      *frostdb.DB
	sql     *gorm.DB
	ts      *Tables
	session *SessionCache
	cancel  context.CancelFunc

	events   chan *Event
	sessions chan *Session
}

func New(ctx context.Context, o *Config) (*Vince, error) {
	sqlPath := filepath.Join(o.DataPath, "sql")
	err := os.MkdirAll(sqlPath, 0755)
	if err != nil {
		return nil, err
	}
	sqlDb, err := open(sqlPath)
	if err != nil {
		return nil, err
	}
	store, err := frostdb.New(
		frostdb.WithStoragePath(o.DataPath),
	)
	if err != nil {
		closeDB(sqlDb)
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	db, err := store.DB(ctx, "vince")
	if err != nil {
		closeDB(sqlDb)
		store.Close()
		return nil, err
	}
	tbl, err := NewTables(db)
	if err != nil {
		closeDB(sqlDb)
		store.Close()
		db.Close()
		return nil, err
	}
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		closeDB(sqlDb)
		store.Close()
		db.Close()
		return nil, err
	}
	v := &Vince{
		store:    store,
		db:       db,
		sql:      sqlDb,
		ts:       tbl,
		cancel:   cancel,
		events:   make(chan *Event, 1024),
		sessions: make(chan *Session, 1024),
	}
	v.session = NewSessionCache(cache, v.sessions)
	return v, nil
}

func (v *Vince) Start() {
	v.pool.Go(v.loopEvent)
	v.pool.Go(v.loopSessions)
}

func (v *Vince) Close() {
	v.cancel()
	v.pool.Wait()
	closeDB(v.sql)
	v.db.Close()
	v.store.Close()
	v.session.cache.Clear()
}

func (v *Vince) loopEvent(ctx context.Context) error {
	events := make(EventList, MAX_BUFFER_SIZE)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-v.events:
			events = append(events, e)
			if len(events) >= MAX_BUFFER_SIZE {
				n, err := events.Save(ctx, v.ts)
				if err != nil {
					xlg.Err(err).Msg("Failed to save events")
				} else {
					xlg.Trace().Uint64("size", n).Msg("saved events")
				}
				for _, ev := range events {
					PutEvent(ev)
				}
				events = events[:0]
			}
		}
	}
}

func (v *Vince) loopSessions(ctx context.Context) error {
	sessions := make(SessionList, MAX_BUFFER_SIZE)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sess := <-v.sessions:
			sessions = append(sessions, sess)
			if len(sessions) >= MAX_BUFFER_SIZE {
				n, err := sessions.Save(ctx, v.ts)
				if err != nil {
					xlg.Err(err).Msg("Failed to save sessions")
				} else {
					xlg.Trace().Uint64("size", n).Msg("saved sessions")
				}
				for _, s := range sessions {
					PutSession(s)
				}
				sessions = sessions[:0]
			}
		}
	}
}

var domainStatusRe = regexp.MustCompile(`^/(?:[a-zA-Z]([a-zA-Z0-9\-]+[\.]?)*[a-zA-Z0-9])/status$`)

func (v *Vince) serveAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		switch r.URL.Path {
		case "/api/events":
			v.EventsEndpoint(w, r)
			return
		case "/subscription/webhook":
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		}
	case http.MethodGet:
		switch r.URL.Path {
		case "/error":
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		case "/health":
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		case "/system":
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
			return
		default:
			if domainStatusRe.Match([]byte(r.URL.Path)) {
				http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
				return
			}
		}
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
