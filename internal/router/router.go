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

	// Pipeline for public facing apis. This includes events ingestion api and
	// health endpoints.
	public := plug.Public(ctx)

	browser := plug.Browser(ctx)

	return plug.Pipeline{
		browser.PathGET("/metrics", metrics.ServeHTTP),
		// add prefix matches on the top of the pipeline for faster lookups
		plug.Ok(config.Get(ctx).EnableProfile,
			browser.Prefix("/debug/pprof/", pprof.Index),
		),

		public.PathPOST("/api/event", api.Events),
		public.PathGET("/health", api.Health),
		public.PathGET("/version", api.Version),

		NotFound,
	}
}

func NotFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}
