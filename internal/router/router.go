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
	pipe4 := plug.API(ctx)
	pipe6 := plug.InternalStatsAPI()
	browser := plug.Browser(ctx)
	pipe5 := append(plug.Browser(ctx), plug.Protect()...)
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

		pipe4.PathPOST("/api/event", api.Events),
		pipe4.PathGET("/health", api.Health),
		pipe4.PathGET("/version", api.Version),

		browser.PathGET("/", pages.Home),
		browser.PathGET("/avatar", avatar.Serve),
		pipe5.And(plug.RequireLoggedOut).PathGET("/register", auth.RegisterForm),
		pipe5.And(plug.RequireLoggedOut).PathPOST("/register", auth.Register),
		pipe5.And(plug.RequireLoggedOut).GET(`^/register/invitation/:invitation_id$`, auth.RegisterFromInvitationForm),
		pipe5.And(plug.RequireLoggedOut).POST(`^/register/invitation/:invitation_id$`, auth.RegisterFromInvitation),
		pipe5.And(plug.RequireAccount).PathGET("/activate", auth.ActivateForm),
		pipe5.PathPOST("/activate", auth.Activate),
		pipe5.PathPOST("/activate/request-code", auth.RequestActivationCode),
		pipe5.PathGET("/login", auth.LoginForm),
		pipe5.And(plug.RequireLoggedOut).PathPOST("/login", auth.Login),
		pipe5.PathGET("/password/request-reset", auth.PasswordResetRequestForm),
		pipe5.PathPOST("/password/request-reset", auth.PasswordResetRequest),
		pipe5.PathGET("/password/reset", auth.PasswordResetForm),
		pipe5.PathPOST("/password/reset", auth.PasswordReset),

		pipe5.And(plug.RequireAccount).PathGET("/password", auth.PasswordForm),
		pipe5.And(plug.RequireAccount).PathPOST("/password", auth.SetPassword),
		pipe5.PathGET("/logout", auth.Logout),
		pipe5.And(plug.RequireAccount).PathGET("/settings", auth.UserSettings),
		pipe5.And(plug.RequireAccount).PathPOST("/settings", auth.SaveSettings),
		pipe5.And(plug.RequireAccount).PathDELETE("/me", auth.DeleteMe),
		pipe5.PathPOST("/settings/tokens", auth.CreatePersonalAccessToken),
		pipe5.DELETE(`^/settings/tokens/:id$`, auth.DeleteAPIKey),

		plug.PREFIX("/sites",
			pipe5.And(plug.RequireAccount).PathPOST("/sites", site.CreateSite),
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

		pipe5.And(plug.RequireAccount).PathGET("/new", site.New),
		sitePipe.GET("/:owner/:site", site.Home),
		pipe5.And(plug.RequireAccount).GET("/:owner", user.Home),
		NotFound,
	}
}

func NotFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.ERROR(r.Context(), w, http.StatusNotFound)
	})
}
