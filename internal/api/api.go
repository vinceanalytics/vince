package api

import (
	"compress/gzip"
	"context"
	"crypto/subtle"
	"io"
	"net/http"
	"strings"
	"sync"

	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/internal/buffers"
	"github.com/vinceanalytics/vince/internal/guard"
	"github.com/vinceanalytics/vince/internal/logger"
	"github.com/vinceanalytics/vince/internal/request"
	"github.com/vinceanalytics/vince/internal/session"
	"github.com/vinceanalytics/vince/internal/stats"
	"github.com/vinceanalytics/vince/internal/tenant"
	"github.com/vinceanalytics/vince/internal/tracker"
	"github.com/vinceanalytics/vince/version"
)

const (
	vary            = "Vary"
	acceptEncoding  = "Accept-Encoding"
	contentEncoding = "Content-Encoding"
	contentType     = "Content-Type"
	contentLength   = "Content-Length"
)

type API struct {
	config  *v1.Config
	tenants *tenant.Tenants
	hand    http.Handler
}

var trackerServer = http.FileServer(http.FS(tracker.JS))

var gzipPool = &sync.Pool{New: func() any {
	return gzip.NewWriter(nil)
}}

const minSizeToCompress = 1 << 10

func getZip() *gzip.Writer {
	return gzipPool.Get().(*gzip.Writer)
}

func putZip(w *gzip.Writer) {
	w.Reset(io.Discard)
	gzipPool.Put(w)
}

func New(ctx context.Context, o *v1.Config, tenants *tenant.Tenants) (*API, error) {
	a := &API{
		config:  o,
		tenants: tenants,
	}
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/") {
			r = r.WithContext(tenants.Load(r.Context(), r.URL.Query()))
		}
		w.Header().Add(vary, acceptEncoding)
		code := &statsWriter{ResponseWriter: w, compress: acceptsGzip(r)}
		defer func() {
			logger.Get(r.Context()).Debug(r.URL.Path, "method", r.Method, "status", code.code)
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
				request.Write(r.Context(), w, &v1.GetDomainResponse{Domains: tenants.Domains(tenant.Default)})
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
			case "/api/v1/stats/breakdown":
				stats.BreakDown(w, r)
				return
			default:
				if strings.HasPrefix(r.URL.Path, "/js/") {
					w.Header().Set("x-content-type-options", "nosniff")
					w.Header().Set("cross-origin-resource-policy", "cross-origin")
					w.Header().Set("access-control-allow-origin", "*")
					w.Header().Set("cache-control", "public, max-age=86400, must-revalidate")
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
				ReceiveEvent(tenants, w, r)
				return
			}
		case http.MethodHead:
			if strings.HasPrefix(r.URL.Path, "/js/") {
				w.Header().Set("x-content-type-options", "nosniff")
				w.Header().Set("cross-origin-resource-policy", "cross-origin")
				w.Header().Set("access-control-allow-origin", "*")
				w.Header().Set("cache-control", "public, max-age=86400, must-revalidate")
				trackerServer.ServeHTTP(w, r)
				return
			}
		}
		request.Error(r.Context(), w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
	})

	a.hand = base
	return a, nil
}

type statsWriter struct {
	http.ResponseWriter
	raw        int
	compressed int
	compress   bool
	code       int
}

func (r *statsWriter) Write(p []byte) (int, error) {
	// All writes to response are a single call.
	r.raw = len(p)
	if !r.compress || len(p) <= minSizeToCompress {
		return r.ResponseWriter.Write(p)
	}
	r.Header().Set(contentEncoding, "gzip")
	r.Header().Del(contentLength)
	if r.code != 0 {
		r.ResponseWriter.WriteHeader(r.code)
	}
	b := buffers.Bytes()
	defer b.Release()
	g := getZip()
	defer putZip(g)
	g.Reset(b)
	g.Write(p)
	g.Close()
	r.compressed = b.Len()
	return r.ResponseWriter.Write(b.Bytes())
}

func (r *statsWriter) WriteHeader(code int) {
	r.code = code
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
	session.Get(ctx).Queue(ctx, tenant.Get(ctx), &req)
	request.Write(ctx, w, &v1.SendEventResponse{Dropped: false})
}
