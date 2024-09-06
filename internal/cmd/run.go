package cmd

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/web"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/db/plug"
	"golang.org/x/crypto/acme/autocert"
)

func run() {
	db, err := db.Open(config.C.DataPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db.Start(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc("/favicon/sources/placeholder", web.Placeholder)
	mux.HandleFunc("/favicon/sources/{source...}", web.Favicon)

	mux.HandleFunc("/{$}", db.Wrap(
		plug.Browser().Then(web.Home),
	))

	mux.HandleFunc("GET /login", db.Wrap(
		plug.Browser().
			With(plug.CSRF).
			Then(web.LoginForm),
	))

	mux.HandleFunc("POST /login", db.Wrap(
		plug.Browser().
			With(plug.VerifyCSRF).
			With(plug.RequireLogout).
			Then(web.Login),
	))

	mux.HandleFunc("GET /logout", db.Wrap(
		plug.Browser().
			Then(web.Logout),
	))

	mux.HandleFunc("GET /sites/new", db.Wrap(
		plug.Browser().
			With(plug.CSRF).
			With(plug.RequireAccount).
			Then(web.CreateSiteForm),
	))

	mux.HandleFunc("POST /sites", db.Wrap(
		plug.Browser().
			With(plug.CSRF).
			With(plug.RequireAccount).
			Then(web.CreateSite),
	))

	mux.HandleFunc("GET /sites", db.Wrap(
		plug.Browser().
			With(plug.RequireAccount).
			Then(web.Sites),
	))

	sites := plug.Browser().
		With(plug.RequireAccount).
		With(web.RequireSiteAccess)

	mux.HandleFunc("GET /v1/share/{domain}", db.Wrap(
		plug.Browser().
			With(web.RequireSiteAccess).
			Then(web.Share),
	))

	mux.HandleFunc("GET /v1/share/{domain}/authenticate/{slug}", db.Wrap(
		plug.Browser().
			With(plug.CSRF).
			Then(web.ShareAuthForm),
	))

	mux.HandleFunc("POST /v1/share/{domain}/authenticate/{slug}", db.Wrap(
		plug.Browser().
			With(plug.CSRF).
			Then(web.ShareAuth),
	))

	mux.HandleFunc("GET /{domain}/shared-links", db.Wrap(
		sites.
			With(plug.CSRF).
			Then(web.SharedLinksForm),
	))

	mux.HandleFunc("GET /shared-links/{domain}/edit/{slug}", db.Wrap(
		sites.
			With(plug.CSRF).
			Then(web.EditSharedLinksForm),
	))

	mux.HandleFunc("POST /shared-links/{domain}/edit/{slug}", db.Wrap(
		sites.
			With(plug.VerifyCSRF).
			Then(web.EditSharedLink),
	))

	mux.HandleFunc("/shared-links/{domain}/delete/{slug}", db.Wrap(
		sites.
			Then(web.DeleteSharedLink),
	))

	mux.HandleFunc("POST /{domain}/shared-links", db.Wrap(
		sites.
			With(plug.VerifyCSRF).
			Then(web.CreateSharedLink),
	))

	mux.HandleFunc("/{domain}/make-public", db.Wrap(
		sites.
			Then(web.MakePublic),
	))

	mux.HandleFunc("/{domain}/make-private", db.Wrap(
		sites.
			Then(web.MakePrivate),
	))

	mux.HandleFunc("GET /{domain}/snippet", db.Wrap(
		sites.
			Then(web.AddSnippet),
	))

	mux.HandleFunc("GET /{domain}/settings", db.Wrap(
		sites.
			With(plug.CSRF).
			Then(web.Settings),
	))

	mux.HandleFunc("/{domain}/delete", db.Wrap(
		sites.
			Then(web.Delete),
	))

	stats := plug.InternalStats().
		With(web.RequireSiteAccess)

	mux.HandleFunc("GET /api/stats/{domain}/current-visitors", db.Wrap(
		stats.
			Then(web.CurrentVisitors),
	))

	mux.HandleFunc("GET /api/stats/{domain}/main-graph/", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /api/stats/{domain}/top-stats/", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /api/stats/{domain}/sources/", db.Wrap(
		stats.
			Then(web.Sources),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_mediums/", db.Wrap(
		stats.
			Then(web.UtmMediums),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_sources/", db.Wrap(
		stats.
			Then(web.UtmSources),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_campaigns/", db.Wrap(
		stats.
			Then(web.UtmCampaigns),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_contents/", db.Wrap(
		stats.
			Then(web.UtmContents),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_terms/", db.Wrap(
		stats.
			Then(web.UtmTerms),
	))

	mux.HandleFunc("GET /api/stats/{domain}/referrers/{referrer}/", db.Wrap(
		stats.
			Then(web.Referrer),
	))

	mux.HandleFunc("GET /api/stats/{domain}/pages/", db.Wrap(
		stats.
			Then(web.Pages),
	))

	mux.HandleFunc("GET /api/stats/{domain}/entry-pages/", db.Wrap(
		stats.
			Then(web.EntryPages),
	))

	mux.HandleFunc("GET /api/stats/{domain}/exit-pages/", db.Wrap(
		stats.
			Then(web.ExitPages),
	))

	mux.HandleFunc("GET /api/stats/{domain}/countries/", db.Wrap(
		stats.
			Then(web.Countries),
	))

	mux.HandleFunc("GET /api/stats/{domain}/regions/", db.Wrap(
		stats.
			Then(web.Regions),
	))

	mux.HandleFunc("GET /api/stats/{domain}/cities/", db.Wrap(
		stats.
			Then(web.Cities),
	))

	mux.HandleFunc("GET /api/stats/{domain}/browsers/", db.Wrap(
		stats.
			Then(web.Browsers),
	))

	mux.HandleFunc("GET /api/stats/{domain}/browser-versions/", db.Wrap(
		stats.
			Then(web.BrowserVersions),
	))

	mux.HandleFunc("GET /api/stats/{domain}/operating-systems/", db.Wrap(
		stats.
			Then(web.Os),
	))

	mux.HandleFunc("GET /api/stats/{domain}/operating-system-versions/", db.Wrap(
		stats.
			Then(web.OsVersion),
	))

	mux.HandleFunc("GET /api/stats/{domain}/screen-sizes/", db.Wrap(
		stats.
			Then(web.ScreenSize),
	))

	mux.HandleFunc("GET /api/stats/{domain}/conversions/", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /api/stats/{domain}/custom-prop-values/{prop_key}/", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /api/stats/{domain}/suggestions/{filter_name}/", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /{domain}/dashboard", db.Wrap(
		plug.Browser().
			With(web.RequireSiteAccess).
			Then(web.Stats),
	))

	mux.HandleFunc("GET /avatar/{size}/{uid...}", db.Wrap(
		plug.Browser().
			With(plug.RequireAccount).
			Then(web.Avatar),
	))

	mux.HandleFunc("/api/event", db.Wrap(web.Event))

	go func() {
		// we load location and ua data async.
		location.GetCity(0)
	}()
	svr := &http.Server{
		Addr:        config.C.Listen,
		BaseContext: func(l net.Listener) context.Context { return ctx },
		Handler: plug.Compress(
			plug.Static(mux),
		),
	}
	if config.C.Acme {
		slog.Info("Auto tls enabled, configuring tls", "email", config.C.AcmeEmail, "domain", config.C.AcmeDomain)
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(config.C.AcmeDomain),
			Email:      config.C.AcmeEmail,
			Cache:      autocert.DirCache(filepath.Join(config.C.DataPath, "certs")),
		}
		svr.TLSConfig = m.TLSConfig()
	}

	slog.Info("starting server", "addr", svr.Addr)
	go func() {
		defer cancel()
		if config.C.Acme {
			svr.ListenAndServeTLS("", "")
		} else {
			svr.ListenAndServe()
		}
	}()
	<-ctx.Done()
	svr.Close()
	slog.Info("Shutting down")
}
