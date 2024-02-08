package api

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/guard"
	"github.com/vinceanalytics/vince/logger"
	"github.com/vinceanalytics/vince/request"
	"github.com/vinceanalytics/vince/session"
	"github.com/vinceanalytics/vince/stats"
	"github.com/vinceanalytics/vince/tracker"
	"github.com/vinceanalytics/vince/version"
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
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := &responseCode{ResponseWriter: w}
		defer func() {
			logger.Get(r.Context()).Debug(r.URL.String(), "method", r.Method, "status", code.code)
		}()
		w = code
		if a.config.AuthToken != "" && r.URL.Path != "/api/event" {
			if subtle.ConstantTimeCompare([]byte(a.config.AuthToken), []byte(parseBearer(r.Header.Get("Authorization")))) != 1 {
				request.Error(r.Context(), w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
				return
			}
		}
		switch r.Method {
		case http.MethodGet:
			switch r.URL.Path {
			case "/api/v1/version":
				request.Write(r.Context(), w, &v1.Version{Version: version.VERSION})
				return
			case "/api/v1/domains":
				o := make([]*v1.Domain, 0, len(a.config.Domains))
				for _, n := range a.config.Domains {
					o = append(o, &v1.Domain{Name: n})
				}
				request.Write(r.Context(), w, &v1.GetDomainResponse{Domains: o})
				return
			case "/api/v1/stats/realtime/visitors":
				stats.Realtime(w, r)
				return
			case "/api/v1/stats/aggregate":
				stats.Aggregate(w, r)
				return
			case "/api/v1/stats/timeseries":
				stats.TimeSeries(w, r)
				return
			default:
				if strings.HasPrefix(r.URL.Path, "/js/") {
					trackerServer.ServeHTTP(w, r)
					return
				}
			}
		case http.MethodPost:
			switch r.URL.Path {
			case "/api/v1/event":
				SendEvent(w, r)
				return
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
		request.Error(r.Context(), w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	})

	a.hand = base
	return a, nil
}

type responseCode struct {
	http.ResponseWriter
	code int
}

func (r *responseCode) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

func parseBearer(auth string) (token string) {
	const prefix = "Bearer "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return ""
	}
	return auth[len(prefix):]
}
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.hand.ServeHTTP(w, r)
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
