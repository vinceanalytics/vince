package server

import (
	"context"
	"fmt"
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
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_protovalidate "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/proto/gen/go/vince/api/v1"
	clusterv1 "github.com/vinceanalytics/proto/gen/go/vince/cluster/v1"
	configv1 "github.com/vinceanalytics/proto/gen/go/vince/config/v1"
	eventsv1 "github.com/vinceanalytics/proto/gen/go/vince/events/v1"
	goalsv1 "github.com/vinceanalytics/proto/gen/go/vince/goals/v1"
	importv1 "github.com/vinceanalytics/proto/gen/go/vince/import/v1"
	queryv1 "github.com/vinceanalytics/proto/gen/go/vince/query/v1"
	sitesv1 "github.com/vinceanalytics/proto/gen/go/vince/sites/v1"
	snippetsv1 "github.com/vinceanalytics/proto/gen/go/vince/snippets/v1"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/b3"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/engine"
	"github.com/vinceanalytics/vince/internal/g"
	"github.com/vinceanalytics/vince/internal/metrics"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/prober"
	"github.com/vinceanalytics/vince/internal/resource"
	"github.com/vinceanalytics/vince/internal/router"
	"github.com/vinceanalytics/vince/internal/scopes"
	"github.com/vinceanalytics/vince/internal/secrets"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/tokens"
	"github.com/vinceanalytics/vince/internal/worker"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpc_health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
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

func Configure(ctx context.Context, o *config.Options) (context.Context, resource.List) {
	slog.SetDefault(config.Logger(o.Env.String(), o.LogLevel).With("server_id", o.ServerId))
	ctx = g.Open(ctx)

	var resources resource.List

	// we start listeners early to make sure we can actually bind to the network.
	// This saves us managing all long running goroutines we start in this process.
	httpListener := must.Must(net.Listen("tcp", o.ListenAddress))(
		"failed binding network address", o.ListenAddress,
	)
	resources = append(resources, httpListener)
	ctx = core.SetHTTPListener(ctx, httpListener)
	ctx = metrics.Open(ctx)

	ctx, dba := db.Open(ctx, o.DbPath)
	resources = append(resources, dba)
	ctx, os := b3.Open(ctx, o.BlocksStore)
	resources = append(resources, os)
	ctx, ts := timeseries.Open(ctx, os, int(o.GetEventsBufferSize()))
	resources = append(resources, ts)
	ctx, eng := engine.Open(ctx)
	resources = append(resources, eng)
	ctx, requests := worker.SetupRequestsBuffer(ctx)
	resources = append(resources, requests)
	return ctx, resources
}

func Run(ctx context.Context, resources resource.List) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	{
		var scheduler *worker.JobScheduler
		ctx, scheduler = worker.OpenScheduler(ctx)
		resources = append(resources, scheduler)
		g.Get(ctx).Go(func() error {
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
	g.Get(ctx).Go(func() error {
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

	g.Get(ctx).Go(func() error {
		defer cancel()
		return msvr.Start()
	})
	g.Get(ctx).Go(func() error {
		// Ensure we close the servers.
		<-ctx.Done()
		slog.Debug("shutting down gracefully")
		return resources.CloseWithGrace(ctx)
	})

	slog.Debug("started serving http traffic", slog.String("address", plainLS.Addr().String()))
	slog.Debug("started serving mysql clients", slog.String("address", msvr.Listener.Addr().String()))
	return g.Get(ctx).Wait()
}

func Handle(ctx context.Context) http.Handler {
	return router.New(ctx)
}

type Vince struct {
	http.Server
	*prober.GRPCProbe
}

func New(ctx context.Context) *Vince {
	v := &Vince{
		GRPCProbe: prober.NewGRPC(),
	}
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
				AuthGRPC(),
				grpc_protovalidate.StreamServerInterceptor(valid),
			)),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				otelgrpc.UnaryServerInterceptor(),
				met.UnaryServerInterceptor(),
				grpc_logging.UnaryServerInterceptor(InterceptorLogger(), logOpts...),
				AuthGRPCUnary(),
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
	clusterv1.RegisterClusterServer(srv, &api.API{})
	eventsv1.RegisterEventsServer(srv, &api.API{})
	importv1.RegisterImportServer(srv, &api.API{})

	routes := Handle(ctx)
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
	metrics.Get(ctx).MustRegister(met,
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

func AuthGRPCUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx, err = authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func authorize(ctx context.Context, method string) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	var m scopes.Scope
	err = m.Parse(method)
	if err != nil {
		return nil, err
	}
	claims, ok := tokens.ValidWithClaims(secrets.Get(ctx), token, m)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}
	ctx = logging.InjectFields(ctx, logging.Fields{"auth.sub", claims.Subject})
	ctx = core.SetAuth(ctx, &configv1.Client_Auth{
		Name:        claims.Subject,
		AccessToken: token,
		ServerId:    config.Get(ctx).ServerId,
	})
	return tokens.Set(ctx, claims), nil
}

func AuthGRPC() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		var err error
		ctx, err = authorize(ctx, info.FullMethod)
		if err != nil {
			return err
		}
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx
		return handler(srv, wrapped)
	}
}
