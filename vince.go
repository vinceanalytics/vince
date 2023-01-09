package vince

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dgraph-io/ristretto"
	"github.com/polarsignals/frostdb"
	"github.com/sourcegraph/conc/pool"
	"github.com/urfave/cli/v2"
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

func ServeCMD() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "starts a server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "port",
				Usage: "port to listen on",
				Value: 8080,
			},
			&cli.PathFlag{
				Name:  "data",
				Usage: "path to data directory",
				Value: ".vince",
			},
		},
		Action: func(ctx *cli.Context) error {
			goCtx := context.Background()
			svr, err := New(goCtx, &Config{DataPath: ctx.Path("path")})
			if err != nil {
				return err
			}
			defer svr.Close()
			return svr.Serve(goCtx, ctx.Int("port"))
		},
	}
}

func New(ctx context.Context, o *Config) (*Vince, error) {
	os.MkdirAll(o.DataPath, 0755)
	sqlPath := filepath.Join(o.DataPath, "vince.db")
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
		pool:     pool.New().WithContext(ctx),
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

func (v *Vince) Serve(ctx context.Context, port int) error {
	v.Start()
	svr := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: v,
	}
	c := make(chan os.Signal, 1)
	go func() {
		err := svr.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			xlg.Err(err).Msg("Exited server")
			c <- os.Interrupt
		}
	}()
	xlg.Info().Msgf("started serving traffic from %s", svr.Addr)
	signal.Notify(c, os.Interrupt)
	sig := <-c
	xlg.Info().Msgf("received signal %s shutting down the server", sig)
	err := svr.Shutdown(ctx)
	if err != nil {
		return err
	}
	return svr.Close()
}

func (v *Vince) Start() {
	v.pool.Go(v.loopEvent)
	v.pool.Go(v.loopSessions)
}

func (v *Vince) Close() {
	v.cancel()
	v.pool.Wait()
	closeDB(v.sql)
	v.store.Close()
	v.session.cache.Close()
	close(v.events)
	close(v.sessions)
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

var domainStatusRe = regexp.MustCompile(`^/(?P<v0>[^.]+)/status$`)

func (v *Vince) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		v.serveAPI(w, r)
		return
	}
}

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
			if domainStatusRe.MatchString(r.URL.Path) {
				domain := domainStatusRe.FindStringSubmatch(r.URL.Path)[1]
				_ = domain
				http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
				return
			}
		}
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
