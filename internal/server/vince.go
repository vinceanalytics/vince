package server

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/assets"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/health"
	"github.com/vinceanalytics/vince/internal/log"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/plug"
	"github.com/vinceanalytics/vince/internal/router"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/worker"
	"golang.org/x/sync/errgroup"
)

func Serve(o *config.Options, x *cli.Context) error {
	ctx, err := config.Load(o, x)
	if err != nil {
		return err
	}
	ctx, resources := Configure(ctx, o)
	if err != nil {
		return err
	}
	return Run(ctx, resources)
}

type ResourceList []io.Closer

func (r ResourceList) Close() error {
	e := make([]error, 0, len(r))
	for i := len(r) - 1; i > 0; i-- {
		e = append(e, r[i].Close())
	}
	return errors.Join(e...)
}

func Configure(ctx context.Context, o *config.Options) (context.Context, ResourceList) {
	log.Level(config.GetLogLevel(o))

	var resources ResourceList

	// we start listeners early to make sure we can actually bind to the network.
	// This saves us managing all long running goroutines we start in this process.
	httpListener := must.Must(net.Listen("tcp", o.ListenAddress))(
		"failed binding network address", o.ListenAddress,
	)
	resources = append(resources, httpListener)
	ctx = core.SetHTTPListener(ctx, httpListener)

	ctx, dba := db.Open(ctx, o.MetaPath)
	resources = append(resources, dba)
	ctx, ts := timeseries.Open(ctx, o.BlocksPath)
	resources = append(resources, ts)

	h := &health.Config{}
	resources = append(resources, h)
	ctx = health.Set(ctx, h)

	// configure http server
	httpSvr := &http.Server{
		Handler:           Handle(ctx),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	ctx = core.SetHTTPServer(ctx, httpSvr)
	resources = append(resources, httpSvr)

	return ctx, resources
}

func Run(ctx context.Context, resources ResourceList) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	var g errgroup.Group
	{
		o := config.Get(ctx)
		g.Go(func() error {
			worker.SaveBuffers(ctx, o.SyncInterval)
			return nil
		})
	}

	plain := core.GetHTTPServer(ctx)
	plainLS := core.GetHTTPListener(ctx)

	g.Go(func() error {
		return plain.Serve(plainLS)
	})
	g.Go(func() error {
		// Ensure we close the servers.
		<-ctx.Done()
		log.Get().Debug().Msg("shutting down gracefully ")
		return resources.Close()
	})
	log.Get().Debug().Str("address", plainLS.Addr().String()).Msg("started serving  http traffic")
	return g.Wait()
}

func Handle(ctx context.Context) http.Handler {
	pipe := append(
		plug.Pipeline{
			plug.Track(),
			assets.Plug(),
			plug.RequestID,
		},
		router.Pipe(ctx)...,
	)
	h := pipe.Pass(plug.NOOP)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
