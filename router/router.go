package router

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/gernest/vince/api"
	"github.com/gernest/vince/auth"
	"github.com/gernest/vince/avatar"
	"github.com/gernest/vince/pages"
	"github.com/gernest/vince/plug"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/share"
	"github.com/gernest/vince/site"
	"github.com/gernest/vince/sites"
)

func Pipe(ctx context.Context) plug.Pipeline {
	pipe0 := plug.Pipeline{plug.Firewall(ctx), plug.AuthorizeStatsAPI}
	pipe1 := plug.Pipeline{plug.Firewall(ctx), plug.AuthorizeSiteAPI}
	pipe2 := plug.SharedLink()
	pipe4 := plug.API(ctx)
	pipe5 := append(plug.Browser(ctx), plug.Protect()...)
	sitePipe := pipe5.And(plug.RequireAccount, plug.AuthorizedSiteAccess("owner", "admin", "super_admin"))
	return plug.Pipeline{
		// add prefix matches on the top of the pipeline for faster lookups
		pipe5.Prefix("/debug/pprof/", pprof.Index),

		pipe0.PathGET("/api/v1/stats", site.Stats),

		plug.PREFIX("/api/v1/sites",
			pipe1.PathPOST("/api/v1/sites", sites.Create),
			pipe1.PathPUT("/api/v1/sites/goals", sites.CreateGoal),
			pipe1.PathPUT("/api/v1/sites/shared-links", sites.FindOrCreateSharedLink),
			pipe1.GET(`^/api/v1/sites/:site_id$`, sites.Get),
			pipe1.DELETE(`^/api/v1/sites/:site_id$`, sites.Delete),
			pipe1.DELETE(`^/api/v1/sites/goals/:goal_id$`, sites.DeleteGoal),
			NotFound,
		),

		plug.PREFIX("/share/",
			pipe2.GET(`^/share/:domain$`, share.SharedLink),
			pipe2.GET(`^/share/:slug/authenticate$`, share.AuthenticateSharedLink),
			NotFound,
		),

		plug.PREFIX("/api/",
			pipe4.PathPOST("/api/event", api.Events),
			pipe4.PathGET("/api/health", api.Health),
			pipe4.PathGET("/api/system", api.Info),
			NotFound,
		),
		pipe5.PathGET("/", pages.Home),
		pipe5.PathGET("/avatar", avatar.Serve),
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
		pipe5.And(plug.RequireAccount).PathGET("/settings/profile", auth.UserSettingsProfile),
		pipe5.And(plug.RequireAccount).PathGET("/settings/admin", auth.UserSettingsAccount),
		pipe5.And(plug.RequireAccount).PathGET("/settings/api-keys", auth.UserSettingsAPIKeys),
		pipe5.And(plug.RequireAccount).PathPUT("/settings", auth.SaveSettings),
		pipe5.And(plug.RequireAccount).PathDELETE("/me", auth.DeleteMe),
		pipe5.PathGET("/settings/api-keys/new", auth.NewAPIKey),
		pipe5.PathPOST("/settings/api-keys", auth.CreateAPIKey),
		pipe5.DELETE(`^/settings/api-keys/:id$`, auth.DeleteAPIKey),

		plug.PREFIX("/sites",
			pipe5.And(plug.RequireAccount).PathGET("/sites", site.Index),
			pipe5.And(plug.RequireAccount).PathGET("/sites/new", site.New),
			pipe5.And(plug.RequireAccount).PathPOST("/sites", site.CreateSite),
			sitePipe.GET(`^/sites/stats/:domain$`, site.Stats),
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
			sitePipe.PUT(`^/sites/:website/shared-links/:slug$`, site.UpdateSharedLink),
			sitePipe.DELETE(`^/sites/:website/shared-links/:slug$`, site.DeleteSharedLink),
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
		sitePipe.GET(`^/:website/settings/general$`, site.SettingsGeneral),
		sitePipe.GET(`^/:website/settings/people$`, site.SettingsPeople),
		sitePipe.GET(`^/:website/settings/visibility$`, site.SettingsVisibility),
		sitePipe.GET(`^/:website/settings/goals$`, site.SettingsGoals),
		sitePipe.GET(`^/:website/settings/email-reports$`, site.SettingsEmailReports),
		sitePipe.GET(`^/:website/settings/danger-zone$`, site.SettingsDangerZone),
		sitePipe.GET(`^/:website/goals/new$`, site.NewGoal),
		sitePipe.POST(`^/:website/goals$`, site.CreateGoal),
		sitePipe.POST(`^/:website/goals/:id/delete$`, site.DeleteGoal),
		sitePipe.PUT(`^/:website/settings$`, site.UpdateSettings),
		sitePipe.POST(`^/:website/delete$`, site.DeleteSite),
		sitePipe.DELETE(`^/:website/stats$`, site.ResetStats),

		sitePipe.GET(`^/:domain/csv$`, site.CsvExport),
		sitePipe.GET(`^/:domain/stats$`, site.Stats),
		NotFound,
	}

}

func NotFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.ERROR(r.Context(), w, http.StatusNotFound)
	})
}
