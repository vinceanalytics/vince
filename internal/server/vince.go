package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"log/slog"

	"github.com/bufbuild/protovalidate-go"
	"github.com/go-chi/cors"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_protovalidate "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/assets"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	goalsv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/goals/v1"
	queryv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/query/v1"
	sitesv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
	snippetsv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/snippets/v1"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/b3"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/engine"
	"github.com/vinceanalytics/vince/internal/ha"
	"github.com/vinceanalytics/vince/internal/metrics"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/plug"
	"github.com/vinceanalytics/vince/internal/prober"
	"github.com/vinceanalytics/vince/internal/router"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/worker"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpc_health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
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
	slog.SetDefault(config.Logger(o.LogLevel).With("id", o.ServerId))

	var resources ResourceList

	// we start listeners early to make sure we can actually bind to the network.
	// This saves us managing all long running goroutines we start in this process.
	httpListener := must.Must(net.Listen("tcp", o.ListenAddress))(
		"failed binding network address", o.ListenAddress,
	)
	resources = append(resources, httpListener)
	ctx = core.SetHTTPListener(ctx, httpListener)

	ctx, dba := db.Open(ctx, o.DbPath)
	resources = append(resources, dba)
	ctx, dbr := db.OpenRaft(ctx, o.RaftPath)
	resources = append(resources, dbr)
	ctx, os := b3.Open(ctx, o.BlocksStore)
	resources = append(resources, os)

	// NOTE: we must open ha before timeseries. This is to allow graceful
	// propagation of block writes when shutting down.
	ctx, hr := ha.Open(ctx)
	resources = append(resources, hr)
	ctx, ts := timeseries.Open(ctx, os, int(o.GetEventsBufferSize()))
	resources = append(resources, ts)
	ctx, eng := engine.Open(ctx)
	resources = append(resources, eng)
	ctx, requests := worker.SetupRequestsBuffer(ctx)
	resources = append(resources, requests)
	return ctx, resources
}

func Run(ctx context.Context, resources ResourceList) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	var g errgroup.Group
	{
		var scheduler *worker.JobScheduler
		ctx, scheduler = worker.OpenScheduler(ctx, &g)
		resources = append(resources, scheduler)
		g.Go(func() error {
			worker.ProcessRequests(ctx)
			return nil
		})
		// analytics data is partitioned daily
		scheduler.Schedule("timeseries", worker.Daily{}, timeseries.Block(ctx))
	}
	o := config.Get(ctx)
	svr := New(ctx)
	resources = append(resources, svr)

	plainLS := core.GetHTTPListener(ctx)
	g.Go(func() error {
		defer cancel()
		if config.IsTLS(o) {
			return svr.ServeTLS(plainLS, o.TlsCertFile, o.TlsKeyFile)
		}
		return svr.Serve(plainLS)
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

func Handle(ctx context.Context, reg *prometheus.Registry) http.Handler {
	pipe := append(
		plug.Pipeline{
			plug.Track(),
			assets.Plug(),
		},
		router.Pipe(ctx, reg)...,
	)
	h := pipe.Pass(plug.NOOP)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

type Vince struct {
	http.Server
	*prober.GRPCProbe
	*prometheus.Registry
}

func New(ctx context.Context) *Vince {
	v := &Vince{
		GRPCProbe: prober.NewGRPC(),
		Registry:  prometheus.NewRegistry(),
	}
	ctx = metrics.Open(ctx, v.Registry)
	o := config.Get(ctx)
	logOpts := []grpc_logging.Option{
		grpc_logging.WithLogOnEvents(grpc_logging.FinishCall),
		grpc_logging.WithLevels(DefaultCodeToLevelGRPC),
	}
	met := grpc_prometheus.NewServerMetrics(
		grpc_prometheus.WithServerHandlingTimeHistogram(),
	)
	valid := must.Must(protovalidate.New())("failed creating proto validator")
	srv := grpc.NewServer(
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				otelgrpc.StreamServerInterceptor(),
				met.StreamServerInterceptor(),
				grpc_logging.StreamServerInterceptor(InterceptorLogger(), logOpts...),
				auth.StreamServerInterceptor(plug.AuthGRPC),
				selector.StreamServerInterceptor(auth.StreamServerInterceptor(plug.AuthGRPCBasic), selector.MatchFunc(plug.AuthGRPCBasicSelector)),
				selector.StreamServerInterceptor(auth.StreamServerInterceptor(plug.AuthGRPC), selector.MatchFunc(plug.AuthGRPCSelector)),
				grpc_protovalidate.StreamServerInterceptor(valid),
			)),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				otelgrpc.UnaryServerInterceptor(),
				met.UnaryServerInterceptor(),
				grpc_logging.UnaryServerInterceptor(InterceptorLogger(), logOpts...),
				selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(plug.AuthGRPCBasic), selector.MatchFunc(plug.AuthGRPCBasicSelector)),
				selector.UnaryServerInterceptor(auth.UnaryServerInterceptor(plug.AuthGRPC), selector.MatchFunc(plug.AuthGRPCSelector)),
				grpc_protovalidate.UnaryServerInterceptor(valid),
			),
		),
	)
	reflection.Register(srv)
	grpc_health.RegisterHealthServer(srv, v.HealthServer())
	v1.RegisterVinceServer(srv, &api.API{})
	sitesv1.RegisterSitesServer(srv, &api.API{})
	queryv1.RegisterQueryServer(srv, &api.API{})
	goalsv1.RegisterGoalsServer(srv, &api.API{})
	snippetsv1.RegisterSnippetsServer(srv, &api.API{})

	routes := Handle(ctx, v.Registry)
	v.Server = http.Server{
		Addr:              o.ListenAddress,
		Handler:           handleGRPC(srv, routes, o.AllowedOrigins),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
	met.InitializeMetrics(srv)
	v.MustRegister(met,
		collectors.NewBuildInfoCollector(),
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	v.Ready()
	v.Healthy()
	return v
}

func handleGRPC(grpcServer *grpc.Server, otherHandler http.Handler, allowedCORSOrigins []string) http.Handler {
	allowAll := false
	if len(allowedCORSOrigins) == 1 && allowedCORSOrigins[0] == "*" {
		allowAll = true
	}
	origins := map[string]struct{}{}
	for _, o := range allowedCORSOrigins {
		origins[o] = struct{}{}
	}
	wrappedGrpc := grpcweb.WrapServer(grpcServer,
		grpcweb.WithAllowNonRootResource(true),
		grpcweb.WithOriginFunc(func(origin string) bool {
			_, found := origins[origin]
			return found || allowAll
		}))

	corsMiddleware := cors.New(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			_, found := origins[origin]
			return found || allowAll
		},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowCredentials: true,
	})

	return corsMiddleware.Handler(h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		otherHandler.ServeHTTP(w, r)
	}), &http2.Server{}))
}

// DefaultCodeToLevelGRPC is the helper mapper that maps gRPC Response codes to log levels.
func DefaultCodeToLevelGRPC(c codes.Code) grpc_logging.Level {
	switch c {
	case codes.Unknown, codes.Unimplemented, codes.Internal, codes.DataLoss:
		return grpc_logging.LevelError
	default:
		return grpc_logging.LevelDebug
	}
}

func InterceptorLogger() grpc_logging.Logger {
	return grpc_logging.LoggerFunc(func(ctx context.Context, lvl grpc_logging.Level, msg string, fields ...any) {
		switch lvl {
		case grpc_logging.LevelDebug:
			slog.Debug(msg, fields...)
		case grpc_logging.LevelInfo:
			slog.Info(msg, fields...)
		case grpc_logging.LevelWarn:
			slog.Warn(msg, fields...)
		case grpc_logging.LevelError:
			slog.Error(msg, fields...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
