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
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

const MAX_BUFFER_SIZE = 4098

type Config struct {
	DataPath string
}

type Vince struct {
	ts      *timeseries.Tables
	sql     *gorm.DB
	session *timeseries.SessionCache

	events   chan *timeseries.Event
	sessions chan *timeseries.Session
	abort    chan os.Signal
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
			svr, err := New(goCtx, &Config{DataPath: ctx.Path("data")})
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
		ts:       ts,
		sql:      sqlDb,
		events:   make(chan *timeseries.Event, MAX_BUFFER_SIZE),
		sessions: make(chan *timeseries.Session, MAX_BUFFER_SIZE),
		abort:    make(chan os.Signal, 1),
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
		go v.loopSessions(ctx, &wg)
		go v.flushSeries(ctx, &wg)
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
	return nil
}

func (v *Vince) loopEvent(ctx context.Context, wg *sync.WaitGroup) {
	xlg.Debug().Str("worker", "event_writer").Msg("start")
	defer func() {
		xlg.Debug().Str("worker", "event_writer").Msg("exit")
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
		xlg.Debug().Str("worker", "event_writer").Int("count", count).Msg("saving events")
		n, err := v.ts.WriteEvents(events)
		if err != nil {
			xlg.Err(err).Str("worker", "event_writer").Msg("saving events")
		}
		xlg.Debug().Str("worker", "event_writer").Int("count", n).Msg("saved events")
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
	xlg.Debug().Str("worker", "session_writer").Msg("start")
	defer func() {
		xlg.Debug().Str("worker", "session_writer").Msg("exit")
		defer wg.Done()
	}()
	sessions := make([]*timeseries.Session, MAX_BUFFER_SIZE)
	sessions = sessions[:0]
	flush := func() {
		count := len(sessions)
		if count == 0 {
			return
		}
		xlg.Debug().Str("worker", "session_writer").Int("count", count).Msg("saving sessions")
		n, err := v.ts.WriteSessions(sessions)
		if err != nil {
			xlg.Err(err).Str("worker", "session_writer").Msg("saving sessions")
			v.exit()
			return
		}
		xlg.Debug().Str("worker", "session_writer").Int("count", n).Msg("saved sessions")
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

func (v *Vince) flushSeries(ctx context.Context, wg *sync.WaitGroup) {
	xlg.Debug().Str("worker", "series_flush").Msg("start")
	defer func() {
		xlg.Debug().Str("worker", "series_flush").Msg("exit")
		defer wg.Done()
	}()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			xlg.Debug().Str("worker", "series_flush").Msg("flushing events")
			err := v.ts.FlushEvents()
			if err != nil {
				xlg.Err(err).Str("worker", "series_flush").Msg("failed flushing events")
				v.exit()
				return
			}
			xlg.Debug().Str("worker", "series_flush").Msg("flushing sessions")
			err = v.ts.FlushSessions()
			if err != nil {
				xlg.Err(err).Str("worker", "series_flush").Msg("failed flushing sessions")
				v.exit()
				return
			}
		}
	}

}

func (v *Vince) exit() {
	v.abort <- os.Interrupt
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
