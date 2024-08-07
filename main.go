package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/vinceanalytics/vince/internal/assert"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/sys"
	"github.com/vinceanalytics/vince/internal/ua"
	"github.com/vinceanalytics/vince/internal/web"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/db/plug"
	"golang.org/x/crypto/acme/autocert"
)

var (
	listenAddress = flag.String("listen", ":8080", "tcp address to bind the server")
	dataPath      = flag.String("data", ".data", "Path to where database data is stored")
	acme          = flag.Bool("acme", false, "Enables auto tls. When used make sure -acme.email and -acme.domain are set")
	acmeEmail     = flag.String("acme.email", "", "Email address to use with lets enctrypt")
	acmeDomain    = flag.String("acme.domain", "", "Domain name to use with lets encrypt")
)

func main() {
	flag.Parse()
	db, err := db.Open(*dataPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	system, err := sys.New(filepath.Join(*dataPath, "sys"))
	assert.Assert(err == nil, "opening sys storage", "err", err)

	defer system.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db.Start(ctx)
	system.Start(ctx)

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

	mux.HandleFunc("GET /register", db.Wrap(
		plug.Browser().
			With(plug.CSRF).
			With(plug.Captcha).
			Then(web.RegisterForm),
	))

	mux.HandleFunc("POST /register", db.Wrap(
		plug.Browser().
			With(plug.VerifyCSRF).
			Then(web.Register),
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

	mux.HandleFunc("POST /{domain}/make-public", db.Wrap(
		sites.
			Then(web.Unimplemented),
	))

	mux.HandleFunc("POST /{domain}/make-private", db.Wrap(
		sites.
			With(plug.VerifyCSRF).
			Then(web.Unimplemented),
	))

	mux.HandleFunc("GET /{domain}/snippet", db.Wrap(
		sites.
			Then(web.Unimplemented),
	))

	mux.HandleFunc("GET /{domain}/settings", db.Wrap(
		sites.
			Then(web.Unimplemented),
	))

	stats := plug.InternalStats().
		With(web.RequireSiteAccess)

	mux.HandleFunc("GET /api/stats/{domain}/current-visitors", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /api/stats/{domain}/main-graph", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /api/stats/{domain}/top-stats", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /api/stats/{domain}/sources", db.Wrap(
		stats.
			Then(web.Sources),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_mediums", db.Wrap(
		stats.
			Then(web.UtmMediums),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_sources", db.Wrap(
		stats.
			Then(web.UtmSources),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_campaigns", db.Wrap(
		stats.
			Then(web.UtmCampaigns),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_contents", db.Wrap(
		stats.
			Then(web.UtmContents),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_terms", db.Wrap(
		stats.
			Then(web.UtmTerms),
	))

	mux.HandleFunc("GET /api/stats/{domain}/referrers/{referrer}", db.Wrap(
		stats.
			Then(web.Referrer),
	))

	mux.HandleFunc("GET /api/stats/{domain}/pages", db.Wrap(
		stats.
			Then(web.Pages),
	))

	mux.HandleFunc("GET /api/stats/{domain}/entry-pages", db.Wrap(
		stats.
			Then(web.EntryPages),
	))

	mux.HandleFunc("GET /api/stats/{domain}/exit-pages", db.Wrap(
		stats.
			Then(web.ExitPages),
	))

	mux.HandleFunc("GET /api/stats/{domain}/countries", db.Wrap(
		stats.
			Then(web.Countries),
	))

	mux.HandleFunc("GET /api/stats/{domain}/regions", db.Wrap(
		stats.
			Then(web.Regions),
	))

	mux.HandleFunc("GET /api/stats/{domain}/cities", db.Wrap(
		stats.
			Then(web.Cities),
	))

	mux.HandleFunc("GET /api/stats/{domain}/browsers", db.Wrap(
		stats.
			Then(web.Browsers),
	))

	mux.HandleFunc("GET /api/stats/{domain}/browser-versions", db.Wrap(
		stats.
			Then(web.BrowserVersions),
	))

	mux.HandleFunc("GET /api/stats/{domain}/operating-systems", db.Wrap(
		stats.
			Then(web.Os),
	))

	mux.HandleFunc("GET /api/stats/{domain}/operating-system-versions", db.Wrap(
		stats.
			Then(web.OsVersion),
	))

	mux.HandleFunc("GET /api/stats/{domain}/screen-sizes", db.Wrap(
		stats.
			Then(web.ScreenSize),
	))

	mux.HandleFunc("GET /api/stats/{domain}/conversions", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /api/stats/{domain}/custom-prop-values/{prop_key}", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /api/stats/{domain}/suggestions/{filter_name}", db.Wrap(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /avatar/{size}/{uid...}", db.Wrap(
		plug.Browser().
			With(plug.RequireAccount).
			Then(web.Avatar),
	))

	mux.HandleFunc("/system/", db.Wrap(
		plug.Browser().
			With(plug.RequireAccount).
			With(web.RequireSuper).
			Then(web.System(system)),
	))

	mux.HandleFunc("/api/event", db.Wrap(web.Event))

	go func() {
		// we load location and ua data async.
		location.GetCity(0)
		ua.Warm()
	}()
	svr := &http.Server{
		Addr:        *listenAddress,
		BaseContext: func(l net.Listener) context.Context { return ctx },
		Handler: sys.HTTP(
			plug.Compress(
				plug.Static(mux),
			),
		),
	}
	if *acme {
		slog.Info("Auto tls enabled, configuring tls", "email", *acmeEmail, "domain", *acmeDomain)
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(*acmeDomain),
			Email:      *acmeEmail,
			Cache:      autocert.DirCache(filepath.Join(*dataPath, "certs")),
		}
		svr.TLSConfig = m.TLSConfig()
	}

	slog.Info("starting server", "addr", svr.Addr)
	go func() {
		defer cancel()
		if *acme {
			svr.ListenAndServeTLS("", "")
		} else {
			svr.ListenAndServe()
		}
	}()
	<-ctx.Done()
	svr.Close()
	slog.Info("Shutting down")
}
