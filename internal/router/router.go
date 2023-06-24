package router

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/auth"
	"github.com/vinceanalytics/vince/internal/avatar"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/pages"
	"github.com/vinceanalytics/vince/internal/plug"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/share"
	"github.com/vinceanalytics/vince/internal/site"
	"github.com/vinceanalytics/vince/internal/sites"
	"github.com/vinceanalytics/vince/internal/stats"
	"github.com/vinceanalytics/vince/internal/user"
)

func Pipe(ctx context.Context) plug.Pipeline {
	metrics := promhttp.Handler()
	pipe1 := plug.Pipeline{plug.Firewall(ctx), plug.AuthorizeSiteAPI}
	pipe2 := plug.SharedLink()

	// Pipeline for public facing apis. This includes events ingestion api and
	// health endpoints.
	public := plug.API(ctx)

	pipe6 := plug.InternalStatsAPI()
	browser := plug.Browser(ctx)
	pipe5 := append(plug.Browser(ctx), plug.Protect()...)

	// This is the pipeline for authorized owned resources accessed from the web
	// browser.
	o := append(plug.Browser(ctx), plug.Protect()...).And(plug.RequireAccount)

	// Pipeline for accessing publicly reachable resources via the web browser.
	www := append(plug.Browser(ctx), plug.Protect()...).And(plug.RequireLoggedOut)

	sitePipe := pipe5.And(plug.RequireAccount, plug.AuthorizedSiteAccess("owner", "admin", "super_admin"))
	return plug.Pipeline{
		browser.PathGET("/metrics", metrics.ServeHTTP),
		// add prefix matches on the top of the pipeline for faster lookups
		plug.Ok(config.Get(ctx).EnableProfile,
			pipe5.Prefix("/debug/pprof/", pprof.Index),
		),
		plug.PREFIX("/api/stats",
			pipe6.GET("/api/stats/:site", stats.Query),
			NotFound,
		),
		plug.PREFIX("/api/v1/sites",
			pipe1.PathPOST("/api/v1/sites", sites.CreateSite),
			pipe1.PathGET("/api/v1/sites", sites.ListSites),
			pipe1.PathPUT("/api/v1/sites/goals", sites.FindOrCreateGoals),
			pipe1.PathPUT("/api/v1/sites/shared-links", sites.FindOrCreateSharedLink),
			pipe1.DELETE(`^/api/v1/sites/goals/:goal_id$`, sites.DeleteGoal),
			pipe1.GET(`^/api/v1/sites/:site$`, sites.GetSite),
			pipe1.DELETE(`^/api/v1/sites/:site$`, sites.DeleteSite),
			NotFound,
		),

		plug.PREFIX("/share/",
			pipe2.GET(`^/share/:site$`, share.SharedLink),
			pipe2.GET(`^/share/:slug/authenticate$`, share.AuthenticateSharedLink),
			NotFound,
		),

		public.PathPOST("/api/event", api.Events),
		public.PathGET("/health", api.Health),
		public.PathGET("/version", api.Version),

		browser.PathGET("/", pages.Home),
		browser.PathGET("/avatar", avatar.Serve),
		www.PathGET("/register", auth.RegisterForm),
		www.PathPOST("/register", auth.Register),
		www.GET(`^/register/invitation/:invitation_id$`, auth.RegisterFromInvitationForm),
		www.POST(`^/register/invitation/:invitation_id$`, auth.RegisterFromInvitation),
		www.PathGET("/activate", auth.ActivateForm),
		www.PathPOST("/activate", auth.Activate),
		www.PathPOST("/activate/request-code", auth.RequestActivationCode),
		www.PathGET("/login", auth.LoginForm),
		www.PathPOST("/login", auth.Login),
		www.PathGET("/password/request-reset", auth.PasswordResetRequestForm),
		www.PathPOST("/password/request-reset", auth.PasswordResetRequest),
		www.PathGET("/password/reset", auth.PasswordResetForm),
		www.PathPOST("/password/reset", auth.PasswordReset),

		o.PathGET("/password", auth.PasswordForm),
		o.PathPOST("/password", auth.SetPassword),
		o.PathGET("/logout", auth.Logout),
		o.PathGET("/settings", auth.UserSettings),
		o.PathPOST("/settings", auth.SaveSettings),
		o.PathDELETE("/me", auth.DeleteMe),
		o.PathPOST("/settings/tokens", auth.CreatePersonalAccessToken),
		o.DELETE(`^/settings/tokens/:id$`, auth.DeleteAPIKey),

		plug.PREFIX("/sites",
			sitePipe.POST(`^/sites/:site/make-public$`, site.MakePublic),
			sitePipe.POST(`^/sites/:site/make-private$`, site.MakePrivate),
			sitePipe.GET(`^/sites/:site/shared-links/new$`, site.NewSharedLink),
			sitePipe.POST(`^/sites/:site/shared-links$`, site.CreateSharedLink),
			sitePipe.GET(`^/sites/:site/shared-links/:slug/edit$`, site.EditSharedLink),
			sitePipe.POST(`^/sites/:site/shared-links/:slug/update$`, site.UpdateSharedLink),
			sitePipe.POST(`^/sites/:site/shared-links/:slug/delete$`, site.DeleteSharedLink),
			sitePipe.GET(`^/sites/:site/memberships/invite$`, site.InviteMemberForm),
			sitePipe.POST(`^/sites/:site/memberships/invite$`, site.InviteMember),
			sitePipe.POST(`^/sites/invitations/:invitation_id/accept$`, site.AcceptInvitation),
			sitePipe.POST(`^/sites/invitations/:invitation_id/reject$`, site.RejectInvitation),
			sitePipe.DELETE(`^/sites/:site/invitations/:invitation_id/reject$`, site.RemoveInvitation),
			sitePipe.PUT(`^/sites/:site/memberships/:id/role/:new_role$`, site.UpdateRole),
			sitePipe.DELETE(`^/sites/:site/memberships/:id$`, site.RemoveMember),
			NotFound,
		),
		sitePipe.GET(`^/:site/snippet$`, site.AddSnippet),
		sitePipe.GET(`^/:site/settings$`, site.Settings),
		sitePipe.GET(`^/:site/goals/new$`, site.NewGoal),
		sitePipe.POST(`^/:site/goals$`, site.CreateGoal),
		sitePipe.POST(`^/:site/goals/:id/delete$`, site.DeleteGoal),
		sitePipe.POST(`^/:site/delete$`, site.DeleteSite),
		sitePipe.DELETE(`^/:site/stats$`, site.ResetStats),

		o.PathPOST("/new", site.CreateSite),
		o.PathGET("/new", site.New),
		o.GET("/:owner/:site/settings", site.Settings),
		o.GET("/:owner/:site", site.Home),
		o.GET("/:owner", user.Home),
		NotFound,
	}
}

func NotFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.ERROR(r.Context(), w, http.StatusNotFound)
	})
}
