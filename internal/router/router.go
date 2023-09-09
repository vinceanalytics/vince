package router

import (
	"context"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vinceanalytics/vince/assets"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/tracker"
)

type Router struct {
	metrics http.Handler
	pprof   http.Handler
}

func New(ctx context.Context, reg *prometheus.Registry) *Router {
	h := &Router{}
	h.metrics = promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	if config.Get(ctx).EnableProfile {
		h.pprof = http.HandlerFunc(pprof.Index)
	} else {
		h.pprof = http.HandlerFunc(NotFound)
	}
	return h
}

func (h *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/js/vince") {
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("cross-origin-resource-policy", "cross-origin")
		w.Header().Set("access-control-allow-origin", "*")
		w.Header().Set("cache-control", "public, max-age=86400, must-revalidate")
		tracker.Serve(w, r)
		return
	}
	if assets.Match(r.URL.Path) {
		assets.FS.ServeHTTP(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/debug/pprof/") {
		h.pprof.ServeHTTP(w, r)
		return
	}
	switch r.URL.Path {
	case "/metrics":
		h.metrics.ServeHTTP(w, r)
		return
	case "/api/event":
		api.Events(w, r)
		return
	}
	NotFound(w, r)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
