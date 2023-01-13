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
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gernest/vince/assets"
	"github.com/polarsignals/frostdb"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

const MAX_BUFFER_SIZE = 4098

type Config struct {
	DataPath string
}

type Vince struct {
	store   *frostdb.ColumnStore
	db      *frostdb.DB
	sql     *gorm.DB
	ts      *Tables
	session *SessionCache

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
		frostdb.WithLogger(GoKit{}),
	)
	if err != nil {
		closeDB(sqlDb)
		return nil, err
	}
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
		events:   make(chan *Event, MAX_BUFFER_SIZE),
		sessions: make(chan *Session, MAX_BUFFER_SIZE),
	}
	v.session = NewSessionCache(cache, v.sessions)
	return v, nil
}

func (v *Vince) Serve(ctx context.Context, port int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	{
		// add all workers
		wg.Add(2)
		go v.loopEvent(ctx, &wg)
		go v.loopSessions(ctx, &wg)
	}

	svr := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: v.Handle(),
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
	err = svr.Close()
	if err != nil {
		return err
	}
	cancel()
	wg.Wait()

	closeDB(v.sql)
	v.store.Close()
	v.session.cache.Close()
	close(v.events)
	close(v.sessions)
	return nil
}

func (v *Vince) loopEvent(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	events := make(EventList, MAX_BUFFER_SIZE)
	events = events[:0]
	flush := func() {
		count := len(events)
		if count == 0 {
			return
		}
		xlg.Debug().Int("count", count).Msg("Saving events")
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
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			flush()
		case e := <-v.events:
			events = append(events, e)
			if len(events) == MAX_BUFFER_SIZE {
				flush()
			}
		}
	}
}

func (v *Vince) loopSessions(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	sessions := make(SessionList, MAX_BUFFER_SIZE)
	sessions = sessions[:0]
	flush := func() {
		count := len(sessions)
		if count == 0 {
			return
		}
		xlg.Debug().Int("count", count).Msg("saving sessions")
		n, err := sessions.Save(ctx, v.ts)
		if err != nil {
			xlg.Err(err).Msg("failed to save sessions")
		} else {
			xlg.Trace().Uint64("size", n).Msg("saved sessions")
		}
		for _, s := range sessions {
			PutSession(s)
		}
		sessions = sessions[:0]
	}
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			flush()
		case sess := <-v.sessions:
			sessions = append(sessions, sess)
			if len(sessions) == MAX_BUFFER_SIZE {
				flush()
			}
		}
	}
}

var domainStatusRe = regexp.MustCompile(`^/(?P<v0>[^.]+)/status$`)

func (v *Vince) Handle() http.Handler {
	asset := assets.Serve()
	csrf := CSRF()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if assets.Match(r.URL.Path) {
			asset.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api") {
			v.api(w, r)
			return
		}
		if isAdminPath(r.URL.Path, r.Method) {
			WriteSecureBrowserHeaders(w)
			if csrf(w, r) {
				v.admin(w, r)
			}
			return
		}
	})
}

type GoKit struct{}

func (GoKit) Log(kv ...interface{}) error {
	fmt.Println(kv...)
	return nil
}
