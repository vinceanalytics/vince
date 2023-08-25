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

	"log/slog"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/assets"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/engine"
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

type shutdown interface {
	Shutdown(context.Context) error
}

func (r ResourceList) CloseWithGrace(ctx context.Context) error {
	e := make([]error, 0, len(r))
	for i := len(r) - 1; i > 0; i-- {
		if shut, ok := r[i].(shutdown); ok {
			e = append(e, shut.Shutdown(ctx))
		} else {
			e = append(e, r[i].Close())
		}
	}
	return errors.Join(e...)
}

func Configure(ctx context.Context, o *config.Options) (context.Context, ResourceList) {
	slog.SetDefault(config.Logger(o.LogLevel))

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
	ctx, ts := timeseries.Open(ctx, o.BlocksPath, int(o.GetEventsBufferSize()))
	resources = append(resources, ts)
	ctx, eng := engine.Open(ctx)
	resources = append(resources, eng)
	ctx, requests := worker.SetupRequestsBuffer(ctx)
	resources = append(resources, requests)

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
		g.Go(func() error {
			worker.ProcessRequests(ctx)
			return nil
		})
	}
	o := config.Get(ctx)
	plain := core.GetHTTPServer(ctx)
	plainLS := core.GetHTTPListener(ctx)
	g.Go(func() error {
		defer cancel()
		if config.IsTLS(o) {
			return plain.ServeTLS(plainLS, o.TlsCertFile, o.TlsKeyFile)
		}
		return plain.Serve(plainLS)
	})

	msvr := must.Must(engine.Listen(ctx))(
		"failed initializing mysql server",
	)
	resources = append(resources, msvr)
	g.Go(func() error {
		defer cancel()
		return msvr.Start()
	})
	g.Go(func() error {
		// Ensure we close the servers.
		<-ctx.Done()
		slog.Debug("shutting down gracefully")
		return resources.CloseWithGrace(ctx)
	})
	slog.Debug("started serving http traffic", slog.String("address", plainLS.Addr().String()))
	slog.Debug("started serving mysql clients", slog.String("address", msvr.Listener.Addr().String()))
	return g.Wait()
}

func Handle(ctx context.Context) http.Handler {
	pipe := append(
		plug.Pipeline{
			plug.Track(),
			assets.Plug(),
		},
		router.Pipe(ctx)...,
	)
	h := pipe.Pass(plug.NOOP)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
