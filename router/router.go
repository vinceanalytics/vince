package router

import (
	"context"
	"net/http"

	"github.com/gernest/vince/api"
	"github.com/gernest/vince/auth"
	"github.com/gernest/vince/billing"
	"github.com/gernest/vince/plug"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/site"
	"github.com/gernest/vince/sites"
	"github.com/gernest/vince/stats"
)

func Pipe(ctx context.Context) plug.Pipeline {
	pipe0 := plug.Pipeline{plug.Firewall(ctx), plug.AuthorizeStatsAPI}
	pipe1 := plug.Pipeline{plug.Firewall(ctx), plug.AuthorizeSiteAPI}
	pipe2 := plug.SharedLink()
	pipe3 := plug.InternalStatsAPI(ctx)
	pipe4 := plug.API(ctx)
	pipe5 := append(plug.Browser(ctx), plug.Protect()...)
	return plug.Pipeline{
		pipe0.PathGET("/api/v1/stats/realtime/visitors", stats.V1RealtimeVisitors),
		pipe0.PathGET("/api/v1/stats/aggregate", stats.V1Aggregate),
		pipe0.PathGET("/api/v1/stats/breakdown", stats.V1Breakdown),
		pipe0.PathGET("/api/v1/stats/timeseries", stats.V1Timeseries),

		pipe1.PathPOST("/api/v1/sites", sites.Create),
		pipe1.PathPUT("/api/v1/sites/goals", sites.CreateGoal),
		pipe1.PathPUT("/api/v1/sites/shared-links", sites.FindOrCreateSharedLink),
		pipe1.GET(`^/api/v1/sites/(?P<site_id>[^.]+)$`, sites.Get),
		pipe1.DELETE(`^/api/v1/sites/(?P<site_id>[^.]+)$`, sites.Delete),
		pipe1.DELETE(`^/api/v1/sites/goals/(?P<goal_id>[^.]+)$`, sites.DeleteGoal),

		pipe2.GET(`^/share/(?P<domain>[^.]+)$`, stats.SharedLink),
		pipe2.GET(`^/share/(?P<slug>[^.]+)/authenticate$`, stats.AuthenticateSharedLink),

		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/current-visitors$`, stats.CurrentVisitors),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/main-graph$`, stats.MainGraph),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/top-stats$`, stats.TopStats),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/sources$`, stats.Sources),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/utm_mediums$`, stats.UTMMediums),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/utm_sources$`, stats.UTMSources),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/utm_campaigns$`, stats.UTMCampaigns),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/utm_contents$`, stats.UTMContents),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/utm_terms$`, stats.UTMTerms),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/referrers/(?P<referrer>[^.]+)$`, stats.ReferrerDrillDown),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/pages$`, stats.Pages),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/entry-pages$`, stats.EntryPages),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/exit-pages$`, stats.ExitPages),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/countries$`, stats.Countries),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/regions$`, stats.Regions),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/cities$`, stats.Cities),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/browsers$`, stats.Browsers),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/browser-versions$`, stats.BrowserVersion),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/operating-systems$`, stats.OperatingSystemVersions),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/operating-system-versions$`, stats.OperatingSystemVersions),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/screen-sizes$`, stats.ScreenSizes),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/conversions$`, stats.Conversions),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/property/(?P<property>[^.]+)$`, stats.PropBreakdown),
		pipe3.GET(`^/api/stats/(?P<domain>[^.]+)/suggestions/(?P<filter_name>[^.]+)$`, stats.FilterSuggestions),

		pipe4.PathGET("/api/event", api.Events),
		pipe4.PathGET("/api/error", api.Error),
		pipe4.PathGET("/api/health", api.Health),
		pipe4.PathGET("/api/system", api.Info),
		pipe4.PathGET("/api/sites", api.Sites),
		pipe4.PathPOST("/api/subscription/webhook", api.SubscriptionWebhook),
		pipe4.GET("`^/api/(?P<domain>[^.]+)/status$`", api.DomainStatus),

		pipe5.PathGET("/", Home),
		pipe5.PathGET("/register", auth.RegisterForm),
		pipe5.PathPOST("/register", auth.Register),
		pipe5.GET(`^/register/invitation/(?P<invitation_id>[^.]+)$`, auth.RegisterFromInvitationForm),
		pipe5.POST(`^/register/invitation/(?P<invitation_id>[^.]+)$`, auth.RegisterFromInvitation),
		pipe5.PathGET("/activate", auth.ActivateForm),
		pipe5.PathPOST("/activate", auth.Activate),
		pipe5.PathPOST("/activate/request-code", auth.RequestActivationCode),
		pipe5.PathGET("/login", auth.LoginForm),
		pipe5.PathPOST("/login", auth.Login),
		pipe5.PathGET("/password/request-reset", auth.PasswordResetRequestForm),
		pipe5.PathPOST("/password/request-reset", auth.PasswordResetRequest),
		pipe5.PathGET("/password/reset", auth.PasswordResetForm),
		pipe5.PathPOST("/password/reset", auth.PasswordReset),
		pipe5.PathPOST("/error_report", SubmitErrorReport),

		pipe5.PathGET("/password", auth.PasswordForm),
		pipe5.PathPOST("/password", auth.SetPassword),
		pipe5.PathGET("/logout", auth.Logout),
		pipe5.PathGET("/settings", auth.UserSettings),
		pipe5.PathPUT("/settings", auth.SaveSettings),
		pipe5.PathDELETE("/me", auth.DeleteMe),
		pipe5.PathGET("/settings/api-keys/new", auth.NewAPIKey),
		pipe5.PathPOST("/settings/api-keys", auth.CreateAPIKey),
		pipe5.DELETE(`^/settings/api-keys/(?P<id>[^.]+)$`, auth.DeleteAPIKey),
		pipe5.PathGET("/auth/google/callback", auth.GoogleAuthCallback),

		pipe5.PathGET("/billing/change-plan", billing.ChangePlanForm),
		pipe5.GET(`^/billing/change-plan/preview/(?P<plan_id>[^.]+)$`, billing.ChangePlanPreview),
		pipe5.POST(`^/billing/change-plan/(?P<new_plan_id>[^.]+)$`, billing.ChangePlan),
		pipe5.PathGET("/billing/upgrade", billing.Upgrade),
		pipe5.GET(`^/billing/upgrade/(?P<plain_id>[^.]+)$`, billing.UpgradeToPlan),
		pipe5.GET(`^/billing/upgrade/enterprise/(?P<plain_id>[^.]+)$`, billing.UpgradeEnterprisePlan),
		pipe5.GET(`^/billing/change-plan/enterprise/(?P<plain_id>[^.]+)$`, billing.ChangeEnterprisePlan),
		pipe5.PathGET("/billing/upgrade-success", billing.UpgradeSuccess),
		pipe5.PathGET("/billing/subscription/ping", billing.PingSubscription),

		pipe5.PathGET("/sites", site.Index),
		pipe5.PathGET("/sites/new", site.New),
		pipe5.PathPOST("/sites", site.CreateSite),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/make-public$`, site.MakePublic),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/make-private$`, site.MakePrivate),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/weekly-report/enable$`, site.EnableWeeklyReport),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/weekly-report/disable$`, site.DisableWeeklyReport),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/weekly-report/recipients$`, site.AddWeeklyReportRecipient),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/weekly-report/recipients/(?P<recipient>[^.]+)$`, site.RemoveWeeklyReportRecipient),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/monthly-report/enable$`, site.EnableMonthlyReport),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/monthly-report/disable$`, site.DisableMonthlyReport),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/monthly-report/recipients$`, site.AddMonthlyReportRecipient),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/monthly-report/recipients/(?P<recipient>[^.]+)$`, site.RemoveMonthlyReportRecipient),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/spike-notification/enable$`, site.EnableSpikeNotification),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/spike-notification/disable$`, site.DisableSpikeNotification),
		pipe5.PUT(`^/sites/(?P<website>[^.]+)/spike-notification$`, site.UpdateSpikeNotification),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/spike-notification/recipients$`, site.AddSpikeNotificationRecipient),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/spike-notification/recipients/(?P<recipient>[^.]+)$`, site.RemoveSpikeNotificationRecipient),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/shared-links/new$`, site.NewSharedLink),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/shared-links$`, site.CreateSharedLink),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/shared-links/(?P<slug>[^.]+)/edit$`, site.EditSharedLink),
		pipe5.PUT(`^/sites/(?P<website>[^.]+)/shared-links/(?P<slug>[^.]+)$`, site.UpdateSharedLink),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/shared-links/(?P<slug>[^.]+)$`, site.DeleteSharedLink),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/custom-domains/(?P<id>[^.]+)$`, site.DeleteCustomDomain),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/memberships/invite$`, site.InviteMemberForm),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/memberships/invite$`, site.InviteMember),
		pipe5.POST(`^/sites/invitations/(?P<invitation_id>[^.]+)/accept$`, site.AcceptInvitation),
		pipe5.POST(`^/sites/invitations/(?P<invitation_id>[^.]+)/reject$`, site.RejectInvitation),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/invitations/(?P<invitation_id>[^.]+)/reject$`, site.RemoveInvitation),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/transfer-ownership$`, site.TransferOwnershipForm),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/transfer-ownership$`, site.TransferOwnership),
		pipe5.PUT(`^/sites/(?P<website>[^.]+)/memberships/(?P<id>[^.]+)/role/(?P<new_role>[^.]+)$`, site.UpdateRole),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/memberships/(?P<id>[^.]+)$`, site.RemoveMember),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/weekly-report/unsubscribe$`, site.WeeklyReport),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/monthly-report/unsubscribe$`, site.MonthlyReport),

		pipe5.GET(`^/(?P<website>[^.]+)/snippet$`, site.AddSnippet),
		pipe5.GET(`^/(?P<website>[^.]+)/settings$`, site.Settings),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/general$`, site.SettingsGeneral),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/people$`, site.SettingsPeople),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/visibility$`, site.SettingsVisibility),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/goals$`, site.SettingsGoals),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/search-console$`, site.SettingsSearchConsole),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/email-reports$`, site.SettingsEmailReports),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/custom-domain$`, site.SettingsCustomDomain),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/danger-zone$`, site.SettingsDangerZone),
		pipe5.GET(`^/(?P<website>[^.]+)/goals/new$`, site.NewGoal),
		pipe5.POST(`^/(?P<website>[^.]+)/goals$`, site.CreateGoal),
		pipe5.DELETE(`^/(?P<website>[^.]+)/goals/(?P<id>[^.]+)$`, site.DeleteGoal),
		pipe5.PUT(`^/(?P<website>[^.]+)/settings$`, site.UpdateSettings),
		pipe5.PUT(`^/(?P<website>[^.]+)/settings/google$`, site.UpdateGoogleAuth),
		pipe5.DELETE(`^/(?P<website>[^.]+)/settings/google-search$`, site.DeleteGoogleAuth),
		pipe5.DELETE(`^/(?P<website>[^.]+)/settings/google-import$`, site.DeleteGoogleAuth),
		pipe5.DELETE(`^/(?P<website>[^.]+)$`, site.DeleteSite),
		pipe5.DELETE(`^/(?P<website>[^.]+)/stats$`, site.ResetStats),
		pipe5.GET(`^/(?P<website>[^.]+)/import/google-analytics/view-id$`, site.ImportFromGoogleViewIdForm),
		pipe5.POST(`^/(?P<website>[^.]+)/import/google-analytics/view-id$`, site.ImportFromGoogleViewId),
		pipe5.GET(`^/(?P<website>[^.]+)/import/google-analytics/user-metric$`, site.ImportFromGoogleUserMetricNotice),
		pipe5.GET(`^/(?P<website>[^.]+)/import/google-analytics/confirm$`, site.ImportFromGoogleConfirm),
		pipe5.POST(`^/(?P<website>[^.]+)/settings/google-import$`, site.ImportFromGoogle),
		pipe5.DELETE(`^/(?P<website>[^.]+)/settings/forget-imported$`, site.ForgetImported),
		pipe5.GET(`^/(?P<domain>[^.]+)/export$`, site.CsvExport),
		pipe5.GET(`^/(?P<domain>[^.]+)/(?P<path>[^.]+)$`, site.Stats),
		NotFound,
	}

}

func Home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

func S501(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}

func S404(w http.ResponseWriter, r *http.Request) {
	render.ERROR(r.Context(), w, http.StatusNotFound)
}

func NotFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		S404(w, r)
	})
}
