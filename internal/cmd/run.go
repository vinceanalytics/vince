package cmd

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/ops"
	"github.com/vinceanalytics/vince/internal/shards"
	"github.com/vinceanalytics/vince/internal/util/acme"
	"github.com/vinceanalytics/vince/internal/util/oracle"
	"github.com/vinceanalytics/vince/internal/web"
	"github.com/vinceanalytics/vince/internal/web/conversions"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/db/plug"
	"github.com/vinceanalytics/vince/internal/web/suggestions"
	"golang.org/x/crypto/acme/autocert"
)

var serve = &cli.Command{
	Name:  "serve",
	Usage: "Starts vince web server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "listen",
			Usage:       "host:port to bind the server",
			Value:       ":8080",
			Sources:     cli.EnvVars("VINCE_LISTEN"),
			Destination: &oracle.Listen,
		},
		&cli.StringFlag{
			Name:        "data",
			Usage:       "directory to store data",
			Sources:     cli.EnvVars("VINCE_DATA"),
			Value:       "vince-data",
			Destination: &oracle.DataPath,
		},
		&cli.BoolFlag{
			Name:        "autoTLS",
			Usage:       "enables automatic tls",
			Sources:     cli.EnvVars("VINCE_AUTO_TLS"),
			Destination: &oracle.Acme.Enabled,
		},
		&cli.StringFlag{
			Name:        "acmeEmail",
			Usage:       "email address for atomatic tls",
			Sources:     cli.EnvVars("VINCE_ACME_EMAIL"),
			Destination: &oracle.Acme.Email,
		},
		&cli.StringFlag{
			Name:        "acmeDomain",
			Usage:       "domain for atomatic tls",
			Sources:     cli.EnvVars("VINCE_ACME_DOMAIN"),
			Destination: &oracle.Acme.Domain,
		},
		&cli.StringFlag{
			Name:        "url",
			Value:       "http://localhost:8080",
			Usage:       "url resolving to this vince instance",
			Sources:     cli.EnvVars("VINCE_URL"),
			Destination: &oracle.Endpoint,
		},
		&cli.StringFlag{
			Name:        "demo",
			Value:       "vinceanalytics.com",
			Usage:       "Website to use as a demo",
			Sources:     cli.EnvVars("VINCE_DEMO_URL"),
			Destination: &oracle.Demo,
		},
		&cli.StringSliceFlag{
			Name:    "domains",
			Usage:   "list of domains to create on startup",
			Sources: cli.EnvVars("VINCE_DOMAINS"),
		},
		&cli.BoolFlag{
			Name:        "profile",
			Usage:       "registrer http profiles on /debug/ path",
			Sources:     cli.EnvVars("VINCE_PROFILE"),
			Destination: &oracle.Profile,
		},
		&cli.StringFlag{
			Name:    "adminName",
			Usage:   "administrator name",
			Sources: cli.EnvVars("VINCE_ADMIN_NAME"),
		},
		&cli.StringFlag{
			Name:    "adminPassword",
			Usage:   "administrator password",
			Sources: cli.EnvVars("VINCE_ADMIN_PASSWORD"),
		},
	},
	Action: run,
}

func run(ctx context.Context, c *cli.Command) error {
	bdb, err := shards.New(oracle.DataPath)
	if err != nil {
		return err
	}
	defer bdb.Close()

	if name, pass := c.String("adminName"), c.String("adminPassword"); name != "" && pass != "" {
		err := ops.CreateAdmin(bdb.Get(), name, pass)
		if err != nil {
			return err
		}
	}
	db, err := db.Open(bdb, c.StringSlice("domains"))
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db.Start(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc("/favicon/sources/placeholder", web.Placeholder)
	mux.HandleFunc("/favicon/sources/{source...}", web.Favicon)

	if oracle.Profile {
		mux.HandleFunc("GET /debug/pprof/{name}", func(w http.ResponseWriter, r *http.Request) {
			name := r.PathValue("name")
			switch name {
			case "profile":
				pprof.Profile(w, r)
			case "symbol":
				pprof.Symbol(w, r)
			case "trace":
				pprof.Trace(w, r)
			default:
				pprof.Index(w, r)
			}
		})
	}
	mux.HandleFunc("/{$}", db.Wrap("query.Home")(
		plug.Browser().Then(web.Home),
	))

	mux.Handle("GET /{domain}/{$}", db.Wrap("query.Stats")(
		plug.Browser().
			With(web.RequireSiteAccess).
			Then(web.Stats),
	))

	mux.HandleFunc("GET /login", db.Wrap("admin.LoginForm")(
		plug.Browser().
			With(plug.CSRF).
			Then(web.LoginForm),
	))

	mux.HandleFunc("POST /login", db.Wrap("admin.Login")(
		plug.Browser().
			With(plug.VerifyCSRF).
			With(plug.RequireLogout).
			Then(web.Login),
	))

	mux.HandleFunc("GET /settings", db.Wrap("admin.Settings")(
		plug.Browser().
			With(plug.CSRF).
			With(plug.RequireAccount).
			Then(web.UserSetting),
	))

	mux.HandleFunc("GET /settings/api-keys/new", db.Wrap("admin.NewAPIKeyForm")(
		plug.Browser().
			With(plug.CSRF).
			With(plug.RequireAccount).
			Then(web.NewApiKey),
	))

	mux.HandleFunc("POST /settings/api-keys", db.Wrap("admin.CreateAPIKey")(
		plug.Browser().
			With(plug.VerifyCSRF).
			With(plug.RequireAccount).
			Then(web.CreateAPiKey),
	))

	mux.HandleFunc("GET /settings/api-keys/{name}", db.Wrap("admin.DeleteAPIKey")(
		plug.Browser().
			With(plug.RequireAccount).
			Then(web.DeleteAPiKey),
	))

	mux.HandleFunc("GET /logout", db.Wrap("admin.Logout")(
		plug.Browser().
			Then(web.Logout),
	))

	mux.HandleFunc("GET /sites/new", db.Wrap("admin.CreateSiteForm")(
		plug.Browser().
			With(plug.CSRF).
			With(plug.RequireAccount).
			Then(web.CreateSiteForm),
	))

	mux.HandleFunc("POST /sites", db.Wrap("admin.CreateSite")(
		plug.Browser().
			With(plug.CSRF).
			With(plug.RequireAccount).
			Then(web.CreateSite),
	))

	mux.HandleFunc("GET /sites", db.Wrap("admin.Sites")(
		plug.Browser().
			With(plug.RequireAccount).
			Then(web.Sites),
	))

	mux.HandleFunc("GET /api/sites", db.Wrap("admin.SitesIndex")(
		plug.Browser().
			With(plug.RequireAccount).
			Then(web.SitesIndex),
	))

	sites := plug.Browser().
		With(plug.RequireAccount).
		With(web.RequireSiteAccess)

	mux.HandleFunc("GET /v1/share/{domain}", db.Wrap("admin.Share")(
		plug.Browser().
			With(web.RequireSiteAccess).
			Then(web.Share),
	))

	mux.HandleFunc("GET /v1/share/{domain}/authenticate/{slug}", db.Wrap("admin.ShareAuthForm")(
		plug.Browser().
			With(plug.CSRF).
			Then(web.ShareAuthForm),
	))

	mux.HandleFunc("POST /v1/share/{domain}/authenticate/{slug}", db.Wrap("admin.ShareAuth")(
		plug.Browser().
			With(plug.CSRF).
			Then(web.ShareAuth),
	))

	mux.HandleFunc("GET /{domain}/shared-links", db.Wrap("admin.SharedLinksForm")(
		sites.
			With(plug.CSRF).
			Then(web.SharedLinksForm),
	))

	mux.HandleFunc("GET /shared-links/{domain}/edit/{slug}", db.Wrap("admin.EditSharedLinksForm")(
		sites.
			With(plug.CSRF).
			Then(web.EditSharedLinksForm),
	))

	mux.HandleFunc("POST /shared-links/{domain}/edit/{slug}", db.Wrap("admin.EditSharedLink")(
		sites.
			With(plug.VerifyCSRF).
			Then(web.EditSharedLink),
	))

	mux.HandleFunc("/shared-links/{domain}/delete/{slug}", db.Wrap("admin.DeleteSharedLink")(
		sites.
			Then(web.DeleteSharedLink),
	))

	mux.HandleFunc("POST /{domain}/shared-links", db.Wrap("admin.CreateSharedLink")(
		sites.
			With(plug.VerifyCSRF).
			Then(web.CreateSharedLink),
	))

	mux.HandleFunc("GET /{domain}/goals/delete", db.Wrap("admin.DeleteGoal")(
		sites.
			Then(web.DeleteGoal),
	))

	mux.HandleFunc("POST /{domain}/goals", db.Wrap("admin.CreateGoal")(
		sites.
			With(plug.VerifyCSRF).
			Then(web.CreateGoal),
	))

	mux.HandleFunc("GET /{domain}/goals/new", db.Wrap("admin.NewGoalForm")(
		sites.
			With(plug.CSRF).
			Then(web.NewGoalForm),
	))

	mux.HandleFunc("/{domain}/make-public", db.Wrap("admin.MakePublic")(
		sites.
			Then(web.MakePublic),
	))

	mux.HandleFunc("/{domain}/make-private", db.Wrap("site.MakePrivate")(
		sites.
			Then(web.MakePrivate),
	))

	mux.HandleFunc("GET /{domain}/snippet", db.Wrap("site.AddSnippet")(
		sites.
			Then(web.AddSnippet),
	))

	mux.HandleFunc("GET /api/{domain}/status", db.Wrap("site.Status")(
		sites.
			Then(web.Status),
	))

	mux.HandleFunc("GET /{domain}/settings", db.Wrap("site.Settings")(
		sites.
			With(plug.CSRF).
			Then(web.Settings),
	))

	mux.HandleFunc("GET /{domain}/settings/goals", db.Wrap("site.GoalSettings")(
		sites.
			Then(web.GoalSettings),
	))

	mux.HandleFunc("/{domain}/delete", db.Wrap("site.Delete")(
		sites.
			Then(web.Delete),
	))

	stats := plug.InternalStats().
		With(web.RequireSiteAccess)

	mux.HandleFunc("GET /api/stats/{domain}/current-visitors", db.Wrap("api.CurrentVisitors")(
		stats.
			Then(web.CurrentVisitors),
	))

	mux.HandleFunc("GET /api/stats/{domain}/main-graph/", db.Wrap("api.MainGraph")(
		stats.
			Then(web.MainGraph),
	))

	mux.HandleFunc("GET /api/stats/{domain}/top-stats/", db.Wrap("api.TopStats")(
		stats.
			Then(web.TopStats),
	))

	mux.HandleFunc("GET /api/stats/{domain}/sources/", db.Wrap("api.Sources")(
		stats.
			Then(web.Sources),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_mediums/", db.Wrap("api.UtmMediums")(
		stats.
			Then(web.UtmMediums),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_sources/", db.Wrap("api.UtmSources")(
		stats.
			Then(web.UtmSources),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_campaigns/", db.Wrap("api.UtmCampaigns")(
		stats.
			Then(web.UtmCampaigns),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_contents/", db.Wrap("api.UtmContents")(
		stats.
			Then(web.UtmContents),
	))

	mux.HandleFunc("GET /api/stats/{domain}/utm_terms/", db.Wrap("api.UtmTerms")(
		stats.
			Then(web.UtmTerms),
	))

	mux.HandleFunc("GET /api/stats/{domain}/referrers/{referrer}/", db.Wrap("api.Referrer")(
		stats.
			Then(web.Referrer),
	))

	mux.HandleFunc("GET /api/stats/{domain}/pages/", db.Wrap("api.Pages")(
		stats.
			Then(web.Pages),
	))

	mux.HandleFunc("GET /api/stats/{domain}/entry-pages/", db.Wrap("api.EntryPages")(
		stats.
			Then(web.EntryPages),
	))

	mux.HandleFunc("GET /api/stats/{domain}/exit-pages/", db.Wrap("api.ExitPages")(
		stats.
			Then(web.ExitPages),
	))

	mux.HandleFunc("GET /api/stats/{domain}/countries/", db.Wrap("api.Countries")(
		stats.
			Then(web.Countries),
	))

	mux.HandleFunc("GET /api/stats/{domain}/regions/", db.Wrap("api.Regions")(
		stats.
			Then(web.Regions),
	))

	mux.HandleFunc("GET /api/stats/{domain}/cities/", db.Wrap("api.Cities")(
		stats.
			Then(web.Cities),
	))

	mux.HandleFunc("GET /api/stats/{domain}/browsers/", db.Wrap("api.Browsers")(
		stats.
			Then(web.Browsers),
	))

	mux.HandleFunc("GET /api/stats/{domain}/browser-versions/", db.Wrap("api.BrowserVersions")(
		stats.
			Then(web.BrowserVersions),
	))

	mux.HandleFunc("GET /api/stats/{domain}/operating-systems/", db.Wrap("api.Os")(
		stats.
			Then(web.Os),
	))

	mux.HandleFunc("GET /api/stats/{domain}/operating-system-versions/", db.Wrap("api.OsVersions")(
		stats.
			Then(web.OsVersion),
	))

	mux.HandleFunc("GET /api/stats/{domain}/screen-sizes/", db.Wrap("api.ScreenSize")(
		stats.
			Then(web.ScreenSize),
	))

	mux.HandleFunc("GET /api/stats/{domain}/conversions/", db.Wrap("api.Conversion")(
		stats.
			Then(conversions.Conversion),
	))

	mux.HandleFunc("GET /api/stats/{domain}/custom-prop-values/{prop_key}/", db.Wrap("api.Props")(
		stats.
			Then(web.UnimplementedStat),
	))

	mux.HandleFunc("GET /api/stats/{domain}/suggestions/{filter_name}/", db.Wrap("api.Suggestions")(
		stats.
			Then(suggestions.Suggest),
	))

	statsAPI := plug.API().With(api.Authorize)

	mux.HandleFunc("GET /api/v1/stats/realtime/visitors", db.Wrap("api.v1.CurrentVisitors")(
		statsAPI.
			Then(api.CurrentVisitors),
	))

	mux.HandleFunc("GET /api/v1/stats/aggregate", db.Wrap("api.v1.Aggregate")(
		statsAPI.
			Then(api.Agggregates),
	))

	mux.HandleFunc("GET /api/v1/stats/breakdown", db.Wrap("api.v1.Breakdown")(
		statsAPI.
			Then(api.Breakdown),
	))

	mux.HandleFunc("GET /api/v1/stats/timeseries", db.Wrap("api.v1.Timeseries")(
		statsAPI.
			Then(api.Timeseries),
	))

	mux.HandleFunc("/api/event", db.Wrap("api.Event")(web.Event))

	svr := &http.Server{
		Addr:        oracle.Listen,
		BaseContext: func(l net.Listener) context.Context { return ctx },
		Handler:     plug.Static(mux),
	}
	if oracle.Acme.Enabled {
		slog.Info("Auto tls enabled, configuring tls", "email", oracle.Acme.Email, "domain", oracle.Acme.Domain)
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(oracle.Acme.Domain),
			Email:      oracle.Acme.Email,
			Cache:      acme.New(db.Pebble()),
		}
		svr.TLSConfig = m.TLSConfig()
	}

	slog.Info("starting server", "addr", svr.Addr)
	go func() {
		defer cancel()
		if oracle.Acme.Enabled {
			svr.ListenAndServeTLS("", "")
		} else {
			svr.ListenAndServe()
		}
	}()
	<-ctx.Done()
	svr.Close()
	slog.Info("Shutting down")
	return nil
}
