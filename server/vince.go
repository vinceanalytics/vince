package server

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/gernest/vince/assets"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/email"
	"github.com/gernest/vince/health"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/pkg/group"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/plug"
	"github.com/gernest/vince/router"
	"github.com/gernest/vince/sessions"
	"github.com/gernest/vince/timeseries"
	"github.com/gernest/vince/userid"
	"github.com/gernest/vince/worker"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

func Serve(ctx *cli.Context) error {
	conf, err := config.Load(ctx)
	if err != nil {
		return err
	}
	goCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// make sure we create log path
	logPath := filepath.Join(conf.DataPath, "logs")
	os.MkdirAll(logPath, 0755)
	errorLog, err := log.NewRotate(logPath)
	if err != nil {
		return err
	}

	xlg := zerolog.New(zerolog.MultiLevelWriter(
		os.Stderr, errorLog,
	)).Level(zerolog.Level(conf.LogLevel)).With().
		Timestamp().Str("env", conf.Env.String()).Logger()
	goCtx = log.Set(goCtx, &xlg)
	if _, err = os.Stat(conf.DataPath); err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(conf.DataPath, 0755)
		if err != nil {
			return err
		}
	}
	conf.LogLevel = config.Config_info
	goCtx = config.Set(goCtx, conf)
	return HTTP(goCtx, conf, errorLog)
}

type resourceList []io.Closer

type resourceFunc func() error

func (r resourceFunc) Close() error {
	return r()
}

func (r resourceList) Close() error {
	e := make([]error, len(r))
	for i, f := range r {
		e[i] = f.Close()
	}
	return errors.Join(e...)
}

func HTTP(ctx context.Context, o *config.Config, errorLog *log.Rotate) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var g errgroup.Group
	ctx = group.Set(ctx, &g)

	ctx = userid.Open(ctx)

	sqlDb, err := models.Open(models.Database(o))
	if err != nil {
		return err
	}
	var resources resourceList
	resources = append(resources, errorLog)
	resources = append(resources, resourceFunc(func() error {
		return models.CloseDB(sqlDb)
	}))

	ctx = models.Set(ctx, sqlDb)
	ctx, ts, err := timeseries.Open(ctx, o.DataPath)
	if err != nil {
		resources.Close()
		return err
	}
	resources = append(resources, ts)
	ctx, err = caches.Open(ctx)
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed to open caches")
		resources.Close()
		return err
	}
	resources = append(resources, resourceFunc(func() error {
		return caches.Close(ctx)
	}))
	mailer, err := email.FromConfig(o)
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed creating mailer")
		resources.Close()
		return err
	}
	resources = append(resources, mailer)
	ctx = email.Set(ctx, mailer)
	session := sessions.NewSession("_vince")
	ctx = sessions.Set(ctx, session)
	var h health.Health
	addHealth := func(x *health.Ping) {
		h = append(h, x)
	}

	h = append(h, health.Base{
		Key:       "database",
		CheckFunc: models.Check,
	})
	{
		// register and start workers
		g.Go(worker.UpdateCacheSites(ctx, addHealth))
		g.Go(worker.LogRotate(ctx, errorLog, addHealth))
		g.Go(worker.SaveTimeseries(ctx, addHealth))
		g.Go(worker.CollectSYstemMetrics(ctx, addHealth))
	}

	resources = append(resources, h)
	ctx = health.Set(ctx, h)

	svr := &http.Server{
		Addr:    o.ListenAddress,
		Handler: Handle(ctx),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
	// We start by shutting down the server before shutting everything else. So we
	// prepend svr for it to be called first.
	resources = append(resourceList{svr}, resources...)

	g.Go(svr.ListenAndServe)
	g.Go(func() error {
		// Ensure we close the server.
		<-ctx.Done()
		log.Get(ctx).Debug().Msg("shutting down gracefully ")
		return resources.Close()
	})
	log.Get(ctx).Debug().Str("address", svr.Addr).Msg("started serving traffic")
	g.Go(func() error {
		abort := make(chan os.Signal, 1)
		signal.Notify(abort, os.Interrupt)
		sig := <-abort
		log.Get(ctx).Info().Msgf("received signal %s shutting down the server", sig)
		cancel()
		return nil
	})
	return g.Wait()
}

func Handle(ctx context.Context) http.Handler {
	pipe := append(
		plug.Pipeline{
			plug.Track(),
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
