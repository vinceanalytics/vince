package router

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/plug"
)

func Pipe(ctx context.Context, reg *prometheus.Registry) plug.Pipeline {
	metrics := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	browser := plug.Browser()
	a := plug.API()
	return plug.Pipeline{
		browser.PathGET("/metrics", metrics.ServeHTTP),
		// add prefix matches on the top of the pipeline for faster lookups
		plug.Ok(config.Get(ctx).EnableProfile,
			browser.Prefix("/debug/pprof/", pprof.Index),
		),
		a.PathPOST("/api/event", api.Events),
		a.PathGET("/version", api.Version),
		NotFound,
	}
}

func NotFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Method, r.URL)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}
