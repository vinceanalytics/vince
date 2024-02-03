package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/bufbuild/protovalidate-go"
	grpc_protovalidate "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/tracker"
	"github.com/vinceanalytics/staples/staples/version"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

type API struct {
	v1.UnsafeStaplesServer
	hand http.Handler
}

var _ v1.StaplesServer = (*API)(nil)

var trackerServer = http.FileServer(http.FS(tracker.JS))

func New(ctx context.Context, o *v1.Config) (*API, error) {
	a := &API{}
	valid, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	svr := grpc.NewServer(
		grpc.StreamInterceptor(
			grpc_protovalidate.StreamServerInterceptor(valid),
		),
		grpc.UnaryInterceptor(
			grpc_protovalidate.UnaryServerInterceptor(valid),
		),
	)
	v1.RegisterStaplesServer(svr, a)
	web := grpcweb.WrapServer(svr,
		grpcweb.WithAllowNonRootResource(true),
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}))
	api := runtime.NewServeMux()
	reflection.Register(svr)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = v1.RegisterStaplesHandlerFromEndpoint(
		ctx, api, o.Listen, opts,
	)
	if err != nil {
		return nil, err
	}
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/event":
			ReceiveEvent(w, r)
			return
		default:
			if strings.HasPrefix(r.URL.Path, "/api/v1/") {
				api.ServeHTTP(w, r)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/js/") {
				trackerServer.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	root := h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			web.ServeHTTP(w, r)
			return
		}
		base.ServeHTTP(w, r)
	}), &http2.Server{})
	a.hand = root
	return a, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.hand.ServeHTTP(w, r)
}

func (a *API) Build(_ context.Context, _ *emptypb.Empty) (*v1.Version, error) {
	return &v1.Version{
		Version: version.VERSION,
	}, nil
}
