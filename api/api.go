package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/bufbuild/protovalidate-go"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/guard"
	"github.com/vinceanalytics/vince/request"
	"github.com/vinceanalytics/vince/session"
	"github.com/vinceanalytics/vince/stats"
	"github.com/vinceanalytics/vince/tracker"
	"github.com/vinceanalytics/vince/version"
	"google.golang.org/protobuf/types/known/emptypb"
)

type API struct {
	config *v1.Config
	hand   http.Handler
}

var trackerServer = http.FileServer(http.FS(tracker.JS))

func New(ctx context.Context, o *v1.Config) (*API, error) {
	a := &API{
		config: o,
	}
	valid, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	ctx = request.With(ctx, valid)
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			switch r.URL.Path {
			case "/v1/version":
				request.Write(r.Context(), w, &v1.Version{Version: version.VERSION})
				return
			case "/v1/domains":
				o := make([]*v1.Domain, 0, len(a.config.Domains))
				for _, n := range a.config.Domains {
					o = append(o, &v1.Domain{Name: n})
				}
				request.Write(r.Context(), w, &v1.GetDomainResponse{Domains: o})
				return
			case "/v1/visitors":
				stats.Realtime(w, r)
			default:
				if strings.HasPrefix(r.URL.Path, "/js/") {
					trackerServer.ServeHTTP(w, r)
					return
				}
			}
		case http.MethodPost:
			switch r.URL.Path {
			case "/v1/event":
				SendEvent(w, r)
				return
			case "/v1/aggregate":
				stats.Aggregate(w, r)
			case "/v1/timeseries":
				stats.TimeSeries(w, r)
			case "/api/event":
				ReceiveEvent(w, r)
				return
			}
		case http.MethodHead:
			if strings.HasPrefix(r.URL.Path, "/js/") {
				trackerServer.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	a.hand = base
	return a, nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.hand.ServeHTTP(w, r)
}

func (a *API) GetVersion(_ context.Context, _ *emptypb.Empty) (*v1.Version, error) {
	return &v1.Version{
		Version: version.VERSION,
	}, nil
}

func SendEvent(w http.ResponseWriter, r *http.Request) {
	var req v1.Event
	if !request.Read(w, r, &req) {
		return
	}
	ctx := r.Context()
	xg := guard.Get(ctx)
	if !xg.Allow() {
		request.Error(ctx, w, http.StatusTooManyRequests, "Limit exceeded")
		return
	}
	if !xg.Accept(req.D) {
		request.Write(ctx, w, &v1.SendEventResponse{Dropped: true})
		return
	}
	session.Get(ctx).Queue(ctx, &req)
	request.Write(ctx, w, &v1.SendEventResponse{Dropped: false})
}
