package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"

	"github.com/apache/arrow/go/v12/arrow/compute"
	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/gernest/vince/assets"
	"github.com/gernest/vince/assets/tracker"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/email"
	"github.com/gernest/vince/health"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/plug"
	"github.com/gernest/vince/router"
	"github.com/gernest/vince/sessions"
	"github.com/gernest/vince/timeseries"
	"github.com/gernest/vince/worker"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

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
	return HTTP(goCtx, conf)
}

func HTTP(ctx context.Context, o *config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sqlPath := filepath.Join(o.DataPath, "vince.db")
	sqlDb, err := models.Open(sqlPath)
	if err != nil {
		return err
	}
	ctx = models.Set(ctx, sqlDb)
	alloc := memory.DefaultAllocator
	ctx = compute.WithAllocator(ctx, alloc)
	ts, err := timeseries.Open(ctx, alloc, o.DataPath, o.DataTtl.AsDuration())
	if err != nil {
		models.CloseDB(sqlDb)
		return err
	}
	ctx = timeseries.Set(ctx, ts)

	ctx, err = caches.Open(ctx)
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed to open caches")
		models.CloseDB(sqlDb)
		ts.Close()
		return err
	}
	m := timeseries.NewMap()
	ctx = timeseries.SetMap(ctx, m)

	mailer, err := email.FromConfig(o)
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed creating mailer")
		models.CloseDB(sqlDb)
		ts.Close()
		caches.Close(ctx)
		return err
	}
	ctx = email.Set(ctx, mailer)
	session := sessions.NewSession("_vince")
	ctx = sessions.Set(ctx, session)
	var h health.Health
	defer func() {
		for _, o := range h {
			o.Close()
		}
	}()
	abort := make(chan os.Signal, 1)
	exit := func() {
		abort <- os.Interrupt
	}
	var wg sync.WaitGroup
	h = append(h, health.Base{
		Key:       "database",
		CheckFunc: models.Check,
	})
	h = append(h,
		worker.UpdateCacheSites(ctx, &wg, exit),
	)
	ctx = compute.SetExecCtx(ctx, compute.DefaultExecCtx())
	ctx = health.Set(ctx, h)

	svr := &http.Server{
		Addr:    o.ListenAddress,
		Handler: Handle(ctx),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		err := svr.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Get(ctx).Err(err).Msg("Exited server")
			exit()
		}
	}()
	log.Get(ctx).Info().Msgf("started serving traffic from %s", svr.Addr)
	signal.Notify(abort, os.Interrupt)
	sig := <-abort
	log.Get(ctx).Info().Msgf("received signal %s shutting down the server", sig)
	err = svr.Shutdown(ctx)
	if err != nil {
		return err
	}
	err = svr.Close()
	if err != nil {
		return err
	}
	cancel()
	wg.Wait()

	models.CloseDB(sqlDb)
	caches.Close(ctx)
	ts.Close()
	mailer.Close()
	close(abort)
	return nil
}

func Handle(ctx context.Context) http.Handler {
	pipe := append(
		plug.Pipeline{
			tracker.Plug(),
			plug.Favicon(plug.DefaultClient),
			assets.Plug(),
			plug.RequestID,
			plug.CORS,
		},
		router.Pipe(ctx)...,
	)
	h := pipe.Pass(plug.NOOP)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
