package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/apache/arrow/go/v12/arrow/compute"
	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/dgraph-io/ristretto"
	"github.com/gernest/vince/assets"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/email"
	"github.com/gernest/vince/health"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/plug"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
	"github.com/gernest/vince/timeseries"
	"github.com/gernest/vince/worker"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

const MAX_BUFFER_SIZE = 4098

type Vince struct {
	ts     *timeseries.Tables
	sql    *gorm.DB
	mailer email.Mailer
	config *config.Config

	events     chan *timeseries.Event
	sessions   chan *timeseries.Session
	abort      chan os.Signal
	computeCtx compute.ExecCtx
	allocator  memory.Allocator
}

func Serve(ctx *cli.Context) error {
	xlg := zerolog.New(os.Stdout).Level(zerolog.InfoLevel)
	conf, err := config.Load(ctx)
	if err != nil {
		return err
	}
	goCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	xlg = xlg.Level(zerolog.Level(conf.LogLevel))
	goCtx = log.Set(goCtx, &xlg)
	if _, err = os.Stat(conf.DataPath); err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(conf.DataPath, 0755)
		if err != nil {
			return err
		}
	}
	conf.LogLevel = config.Config_info
	err = conf.WriteToFile(filepath.Join(conf.DataPath, "config.json"))
	if err != nil {
		return err
	}
	goCtx = config.Set(goCtx, conf)
	svr, err := New(goCtx, conf)
	if err != nil {
		return err
	}
	return svr.Serve(goCtx)
}

func New(ctx context.Context, o *config.Config) (*Vince, error) {
	sqlPath := filepath.Join(o.DataPath, "vince.db")
	sqlDb, err := models.Open(sqlPath)
	if err != nil {
		return nil, err
	}
	alloc := memory.DefaultAllocator
	ts, err := timeseries.Open(ctx, alloc, o.DataPath)
	if err != nil {
		models.CloseDB(sqlDb)
		return nil, err
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed creating timeseries session cache")
		models.CloseDB(sqlDb)
		ts.Close()
		return nil, err
	}
	mailer, err := email.FromConfig(o)
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed creating mailer")
		models.CloseDB(sqlDb)
		ts.Close()
		return nil, err
	}
	v := &Vince{
		ts:         ts,
		sql:        sqlDb,
		mailer:     mailer,
		config:     o,
		events:     make(chan *timeseries.Event, MAX_BUFFER_SIZE),
		sessions:   make(chan *timeseries.Session, MAX_BUFFER_SIZE),
		abort:      make(chan os.Signal, 1),
		computeCtx: compute.DefaultExecCtx(),
		allocator:  alloc,
	}
	v.ts.Cache = timeseries.NewSessionCache(cache, v.sessions, v.events)
	return v, nil
}

func (v *Vince) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	session := sessions.NewSession("vince")
	ctx = sessions.Set(ctx, session)
	var h health.Health
	defer func() {
		for _, o := range h {
			o.Clone()
		}
	}()
	var wg sync.WaitGroup
	{
		// add all workers
		h = append(h,
			worker.Flush(ctx, "events_worker", v.events, v.ts.WriteEvents, &wg, v.exit),
		)
		h = append(h,
			worker.Flush(ctx, "session_worker", v.sessions, v.ts.WriteSessions, &wg, v.exit),
		)
		h = append(h, worker.StartSeriesArchive(ctx, v.ts, &wg, v.exit))
	}
	h = append(h, health.Base{
		Key:       "database",
		CheckFunc: models.Check,
	})

	svr := &http.Server{
		Addr:    fmt.Sprintf(":%d", v.config.Port),
		Handler: v.Handle(),
		BaseContext: func(l net.Listener) context.Context {
			ctx = compute.SetExecCtx(ctx, v.computeCtx)
			ctx = compute.WithAllocator(ctx, v.allocator)
			ctx = models.Set(ctx, v.sql)
			ctx = email.Set(ctx, v.mailer)
			ctx = timeseries.Set(ctx, v.ts)
			ctx = health.Set(ctx, h)
			return ctx
		},
	}
	go func() {
		err := svr.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Get(ctx).Err(err).Msg("Exited server")
			v.exit()
		}
	}()
	log.Get(ctx).Info().Msgf("started serving traffic from %s", svr.Addr)
	signal.Notify(v.abort, os.Interrupt)
	sig := <-v.abort
	log.Get(ctx).Info().Msgf("received signal %s shutting down the server", sig)

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

	models.CloseDB(v.sql)
	v.ts.Close()
	v.mailer.Close()
	close(v.events)
	close(v.sessions)
	close(v.abort)
	return nil
}

func (v *Vince) exit() {
	v.abort <- os.Interrupt
}

var domainStatusRe = regexp.MustCompile(`^/(?P<v0>[^.]+)/status$`)

func (v *Vince) Handle() http.Handler {
	asset := assets.Serve()
	admin := plug.Chain(Admin(), append(plug.Browser(), plug.SecureForm()...)...)
	home := v.home()
	v1Stats := v1Stats()
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
		render.ERROR(r.Context(), w, http.StatusNotFound)
	})
}
