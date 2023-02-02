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
	"github.com/gernest/vince/timeseries"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

const MAX_BUFFER_SIZE = 4098

type Config struct {
	DataPath      string
	FlushInterval time.Duration
}

type Vince struct {
	ts      *timeseries.Tables
	sql     *gorm.DB
	session *timeseries.SessionCache

	events        chan *timeseries.Event
	sessions      chan *timeseries.Session
	abort         chan os.Signal
	hs            *WorkerHealthChannels
	clientSession *Session
	flushInterval time.Duration
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
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "sets log level to debug",
			},
			&cli.DurationFlag{
				Name:  "flush-interval",
				Value: 30 * time.Minute,
			},
		},
		Action: func(ctx *cli.Context) error {
			goCtx := context.Background()
			if ctx.Bool("debug") {
				setDebug()
			}
			svr, err := New(goCtx, &Config{
				DataPath:      ctx.Path("data"),
				FlushInterval: ctx.Duration("flush-interval"),
			})
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
	ts, err := timeseries.Open(o.DataPath)
	if err != nil {
		closeDB(sqlDb)
		return nil, err
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		closeDB(sqlDb)
		ts.Close()
		return nil, err
	}
	v := &Vince{
		ts:            ts,
		sql:           sqlDb,
		events:        make(chan *timeseries.Event, MAX_BUFFER_SIZE),
		sessions:      make(chan *timeseries.Session, MAX_BUFFER_SIZE),
		abort:         make(chan os.Signal, 1),
		hs:            newWorkerHealth(),
		clientSession: NewSession("vince"),
		flushInterval: o.FlushInterval,
	}
	v.session = timeseries.NewSessionCache(cache, v.sessions)
	return v, nil
}

func (v *Vince) Serve(ctx context.Context, port int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	{
		// add all workers
		wg.Add(3)
		go v.loopEvent(ctx, &wg)
		go v.sessionWriter(ctx, &wg)
		go v.seriesArchive(ctx, &wg)
	}

	svr := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: v.Handle(),
	}
	go func() {
		err := svr.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			xlg.Err(err).Msg("Exited server")
			v.exit()
		}
	}()
	xlg.Info().Msgf("started serving traffic from %s", svr.Addr)
	signal.Notify(v.abort, os.Interrupt)
	sig := <-v.abort
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
	v.ts.Close()
	v.session.Close()
	close(v.events)
	close(v.sessions)
	close(v.abort)
	v.hs.Close()
	return nil
}

func (v *Vince) loopEvent(ctx context.Context, wg *sync.WaitGroup) {
	ev := func() *zerolog.Event {
		return xlg.Debug().Str("worker", "event_writer")
	}
	ev().Msg("start")
	defer func() {
		ev().Msg("exit")
		defer wg.Done()
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	events := make([]*timeseries.Event, MAX_BUFFER_SIZE)
	events = events[:0]
	flush := func() {
		count := len(events)
		if count == 0 {
			return
		}
		_, err := v.ts.WriteEvents(events)
		if err != nil {
			ev().Err(err).Msg("saving events")
			v.exit()
			return
		}
		events = events[:0]
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			flush()
		case ch := <-v.hs.eventWriter:
			ch <- struct{}{}
		case e := <-v.events:
			events = append(events, e)
			if len(events) == MAX_BUFFER_SIZE {
				flush()
			}
		}
	}
}

func (v *Vince) sessionWriter(ctx context.Context, wg *sync.WaitGroup) {
	ev := func() *zerolog.Event {
		return xlg.Debug().Str("worker", "session_writer")
	}
	ev().Msg("start")
	defer func() {
		ev().Msg("exit")
		defer wg.Done()
	}()
	sessions := make([]*timeseries.Session, MAX_BUFFER_SIZE)
	sessions = sessions[:0]
	flush := func() {
		count := len(sessions)
		if count == 0 {
			return
		}
		_, err := v.ts.WriteSessions(sessions)
		if err != nil {
			ev().Err(err).Msg("saving sessions")
			v.exit()
			return
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
		case ch := <-v.hs.sessionWriter:
			ch <- struct{}{}
		case sess := <-v.sessions:
			sessions = append(sessions, sess)
			if len(sessions) == MAX_BUFFER_SIZE {
				flush()
			}
		}
	}
}

func (v *Vince) seriesArchive(ctx context.Context, wg *sync.WaitGroup) {
	ev := func() *zerolog.Event {
		return xlg.Debug().Str("worker", "series_archive")
	}
	ev().Dur("interval", v.flushInterval).Msg("start")
	defer func() {
		ev().Msg("exit")
		defer wg.Done()
	}()
	ticker := time.NewTicker(v.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case ch := <-v.hs.seriesFlush:
			ch <- struct{}{}
		case <-ticker.C:
			n, err := v.ts.ArchiveEvents()
			if err != nil {
				ev().Msg("failed archiving events")
				v.exit()
				return
			}
			ev().Int64("size", n).Msg("archiving events")
			n, err = v.ts.ArchiveSessions()
			if err != nil {
				ev().Err(err).Msg("failed archiving sessions")
				v.exit()
				return
			}
			ev().Int64("size", n).Msg("archiving sessions")
		}
	}

}

func (v *Vince) exit() {
	v.abort <- os.Interrupt
}

var domainStatusRe = regexp.MustCompile(`^/(?P<v0>[^.]+)/status$`)

func (v *Vince) Handle() http.Handler {
	asset := assets.Serve()
	admin := secureBrowser(
		v.csrf(v.admin()),
	)
	home := v.home()
	v1Stats := v.v1Stats()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == http.MethodGet {
			home.ServeHTTP(w, r)
			return
		}
		if assets.Match(r.URL.Path) {
			asset.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/v1/stats") {
			v1Stats.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api") {
			v.api(w, r)
			return
		}

		if isAdminPath(r.URL.Path, r.Method) {
			admin.ServeHTTP(w, r)
			return
		}
		ServeError(w, http.StatusNotFound)
	})
}
