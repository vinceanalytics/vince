package router

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/plug"
)

func Pipe(ctx context.Context) plug.Pipeline {
	metrics := promhttp.Handler()

	browser := plug.Browser()
	a := plug.API()

	return plug.Pipeline{
		browser.PathGET("/metrics", metrics.ServeHTTP),
		// add prefix matches on the top of the pipeline for faster lookups
		plug.Ok(config.Get(ctx).EnableProfile,
			browser.Prefix("/debug/pprof/", pprof.Index),
		),
		a.PathPOST("/api/event", api.Events),
		a.PathGET("/health", api.Health),
		a.PathGET("/version", api.Version),
		a.PathGET("/sites", api.ListSites),
		a.PathPOST("/sites", api.Create),
		NotFound,
	}
}

func NotFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}
