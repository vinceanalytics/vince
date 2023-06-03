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
			pipe6.GET("/api/stats/:website", stats.Query),
			NotFound,
		),

		plug.PREFIX("/api/v1/sites",
			pipe1.PathPOST("/api/v1/sites", sites.CreateSite),
			pipe1.PathGET("/api/v1/sites", sites.ListSites),
			pipe1.PathPUT("/api/v1/sites/goals", sites.FindOrCreateGoals),
			pipe1.PathPUT("/api/v1/sites/shared-links", sites.FindOrCreateSharedLink),
			pipe1.DELETE(`^/api/v1/sites/goals/:goal_id$`, sites.DeleteGoal),
			pipe1.GET(`^/api/v1/sites/:site_id$`, sites.GetSite),
			pipe1.DELETE(`^/api/v1/sites/:site_id$`, sites.DeleteSite),
			NotFound,
		),

		plug.PREFIX("/share/",
			pipe2.GET(`^/share/:domain$`, share.SharedLink),
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
		pipe5.And(plug.RequireAccount).PathPUT("/settings", auth.SaveSettings),
		pipe5.And(plug.RequireAccount).PathDELETE("/me", auth.DeleteMe),
		pipe5.PathPOST("/settings/api-keys", auth.CreateAPIKey),
		pipe5.POST(`^/settings/api-keys/:id/delete$`, auth.DeleteAPIKey),

		plug.PREFIX("/sites",
			pipe5.And(plug.RequireAccount).PathGET("/sites", site.Index),
			pipe5.And(plug.RequireAccount).PathGET("/sites/new", site.New),
			pipe5.And(plug.RequireAccount).PathPOST("/sites", site.CreateSite),
			sitePipe.POST(`^/sites/:website/make-public$`, site.MakePublic),
			sitePipe.POST(`^/sites/:website/make-private$`, site.MakePrivate),
			sitePipe.POST(`^/sites/:website/weekly-report/enable$`, site.EnableWeeklyReport),
			sitePipe.POST(`^/sites/:website/weekly-report/disable$`, site.DisableWeeklyReport),
			sitePipe.POST(`^/sites/:website/weekly-report/recipients$`, site.AddWeeklyReportRecipient),
			sitePipe.DELETE(`^/sites/:website/weekly-report/recipients/:recipient$`, site.RemoveWeeklyReportRecipient),
			sitePipe.POST(`^/sites/:website/monthly-report/enable$`, site.EnableMonthlyReport),
			sitePipe.POST(`^/sites/:website/monthly-report/disable$`, site.DisableMonthlyReport),
			sitePipe.POST(`^/sites/:website/monthly-report/recipients$`, site.AddMonthlyReportRecipient),
			sitePipe.DELETE(`^/sites/:website/monthly-report/recipients/:recipient$`, site.RemoveMonthlyReportRecipient),
			sitePipe.POST(`^/sites/:website/spike-notification/enable$`, site.EnableSpikeNotification),
			sitePipe.POST(`^/sites/:website/spike-notification/disable$`, site.DisableSpikeNotification),
			sitePipe.PUT(`^/sites/:website/spike-notification$`, site.UpdateSpikeNotification),
			sitePipe.POST(`^/sites/:website/spike-notification/recipients$`, site.AddSpikeNotificationRecipient),
			sitePipe.DELETE(`^/sites/:website/spike-notification/recipients/:recipient$`, site.RemoveSpikeNotificationRecipient),
			sitePipe.GET(`^/sites/:website/shared-links/new$`, site.NewSharedLink),
			sitePipe.POST(`^/sites/:website/shared-links$`, site.CreateSharedLink),
			sitePipe.GET(`^/sites/:website/shared-links/:slug/edit$`, site.EditSharedLink),
			sitePipe.POST(`^/sites/:website/shared-links/:slug/update$`, site.UpdateSharedLink),
			sitePipe.POST(`^/sites/:website/shared-links/:slug/delete$`, site.DeleteSharedLink),
			sitePipe.GET(`^/sites/:website/memberships/invite$`, site.InviteMemberForm),
			sitePipe.POST(`^/sites/:website/memberships/invite$`, site.InviteMember),
			sitePipe.POST(`^/sites/invitations/:invitation_id/accept$`, site.AcceptInvitation),
			sitePipe.POST(`^/sites/invitations/:invitation_id/reject$`, site.RejectInvitation),
			sitePipe.DELETE(`^/sites/:website/invitations/:invitation_id/reject$`, site.RemoveInvitation),
			sitePipe.GET(`^/sites/:website/transfer-ownership$`, site.TransferOwnershipForm),
			sitePipe.POST(`^/sites/:website/transfer-ownership$`, site.TransferOwnership),
			sitePipe.PUT(`^/sites/:website/memberships/:id/role/:new_role$`, site.UpdateRole),
			sitePipe.DELETE(`^/sites/:website/memberships/:id$`, site.RemoveMember),
			sitePipe.GET(`^/sites/:website/weekly-report/unsubscribe$`, site.WeeklyReport),
			sitePipe.GET(`^/sites/:website/monthly-report/unsubscribe$`, site.MonthlyReport),
			NotFound,
		),

		sitePipe.GET(`^/:website/snippet$`, site.AddSnippet),
		sitePipe.GET(`^/:website/settings$`, site.Settings),
		sitePipe.GET(`^/:website/goals/new$`, site.NewGoal),
		sitePipe.POST(`^/:website/goals$`, site.CreateGoal),
		sitePipe.POST(`^/:website/goals/:id/delete$`, site.DeleteGoal),
		sitePipe.PUT(`^/:website/settings$`, site.UpdateSettings),
		sitePipe.POST(`^/:website/delete$`, site.DeleteSite),
		sitePipe.DELETE(`^/:website/stats$`, site.ResetStats),
		sitePipe.GET(`^/:domain/stats$`, site.Stats),
		NotFound,
	}

}

func NotFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.ERROR(r.Context(), w, http.StatusNotFound)
	})
}
