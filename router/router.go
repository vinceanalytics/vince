package router

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/gernest/vince/api"
	"github.com/gernest/vince/auth"
	"github.com/gernest/vince/avatar"
	"github.com/gernest/vince/billing"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/pages"
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
	sitePipe := pipe5.And(plug.RequireAccount, plug.AuthorizedSiteAccess("owner", "admin", "super_admin"))
	return plug.Pipeline{
		Load(config.Get(ctx)),
		// add prefix matches on the top of the pipeline for faster lookups
		pipe5.Prefix("/debug/pprof/", pprof.Index),

		plug.PREFIX("/api/v1/stats/",
			pipe0.PathGET("/api/v1/stats/realtime/visitors", stats.V1RealtimeVisitors),
			pipe0.PathGET("/api/v1/stats/aggregate", stats.V1Aggregate),
			pipe0.PathGET("/api/v1/stats/breakdown", stats.V1Breakdown),
			pipe0.PathGET("/api/v1/stats/timeseries", stats.V1Timeseries),
			NotFound,
		),

		plug.PREFIX("/api/v1/sites",
			pipe1.PathPOST("/api/v1/sites", sites.Create),
			pipe1.PathPUT("/api/v1/sites/goals", sites.CreateGoal),
			pipe1.PathPUT("/api/v1/sites/shared-links", sites.FindOrCreateSharedLink),
			pipe1.GET(`^/api/v1/sites/(?P<site_id>[^.]+)$`, sites.Get),
			pipe1.DELETE(`^/api/v1/sites/(?P<site_id>[^.]+)$`, sites.Delete),
			pipe1.DELETE(`^/api/v1/sites/goals/(?P<goal_id>[^.]+)$`, sites.DeleteGoal),
			NotFound,
		),

		plug.PREFIX("/share/",
			pipe2.GET(`^/share/(?P<domain>[^.]+)$`, stats.SharedLink),
			pipe2.GET(`^/share/(?P<slug>[^.]+)/authenticate$`, stats.AuthenticateSharedLink),
			NotFound,
		),

		plug.PREFIX("/api/stats/",
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
		),
		plug.PREFIX("/api/",
			pipe4.PathPOST("/api/event", api.Events),
			pipe4.PathGET("/api/error", api.Error),
			pipe4.PathGET("/api/health", api.Health),
			pipe4.PathGET("/api/system", api.Info),
			pipe4.PathGET("/api/sites", api.Sites),
			pipe4.PathPOST("/api/subscription/webhook", api.SubscriptionWebhook),
			pipe4.GET("`^/api/(?P<domain>[^.]+)/status$`", api.DomainStatus),
			NotFound,
		),

		pipe5.PathGET("/", pages.Home),
		pipe5.PathGET("/pricing", pages.Pricing),
		pipe5.PathGET("/about", pages.Mark("about", "about us")),
		pipe5.PathGET("/creators", pages.Mark("creators", "web analytics for digital creators")),
		pipe5.PathGET("/agencies", pages.Mark("agencies", "web analytics for freelancers and agencies")),
		pipe5.PathGET("/ecommerce", pages.Mark("ecommerce", "web analytics for ecommerce")),
		pipe5.PathGET("/contact", pages.Mark("contact", "contact us")),
		pipe5.PathGET("/privacy", pages.Mark("privacy", "privacy policy")),
		pipe5.PathGET("/data", pages.Mark("data", "data policy")),
		pipe5.PathGET("/terms", pages.Mark("terms", "Terms and Conditions")),
		pipe5.PathGET("/avatar", avatar.Serve),
		pipe5.And(plug.RequireLoggedOut).PathGET("/register", auth.RegisterForm),
		pipe5.And(plug.RequireLoggedOut).PathPOST("/register", auth.Register),
		pipe5.And(plug.RequireLoggedOut).GET(`^/register/invitation/(?P<invitation_id>[^.]+)$`, auth.RegisterFromInvitationForm),
		pipe5.And(plug.RequireLoggedOut).POST(`^/register/invitation/(?P<invitation_id>[^.]+)$`, auth.RegisterFromInvitation),
		pipe5.And(plug.RequireAccount).PathGET("/activate", auth.ActivateForm),
		pipe5.PathPOST("/activate", auth.Activate),
		pipe5.PathPOST("/activate/request-code", auth.RequestActivationCode),
		pipe5.PathGET("/login", auth.LoginForm),
		pipe5.And(plug.RequireLoggedOut).PathPOST("/login", auth.Login),
		pipe5.PathGET("/password/request-reset", auth.PasswordResetRequestForm),
		pipe5.PathPOST("/password/request-reset", auth.PasswordResetRequest),
		pipe5.PathGET("/password/reset", auth.PasswordResetForm),
		pipe5.PathPOST("/password/reset", auth.PasswordReset),
		pipe5.PathPOST("/error_report", SubmitErrorReport),

		pipe5.And(plug.RequireAccount).PathGET("/password", auth.PasswordForm),
		pipe5.And(plug.RequireAccount).PathPOST("/password", auth.SetPassword),
		pipe5.PathGET("/logout", auth.Logout),
		pipe5.And(plug.RequireAccount).PathGET("/settings", auth.UserSettings),
		pipe5.And(plug.RequireAccount).PathPUT("/settings", auth.SaveSettings),
		pipe5.And(plug.RequireAccount).PathDELETE("/me", auth.DeleteMe),
		pipe5.PathGET("/settings/api-keys/new", auth.NewAPIKey),
		pipe5.PathPOST("/settings/api-keys", auth.CreateAPIKey),
		pipe5.DELETE(`^/settings/api-keys/(?P<id>[^.]+)$`, auth.DeleteAPIKey),
		pipe5.PathGET("/auth/google/callback", auth.GoogleAuthCallback),

		plug.PREFIX("/billing/",
			pipe5.PathGET("/billing/change-plan", billing.ChangePlanForm),
			pipe5.GET(`^/billing/change-plan/preview/(?P<plan_id>[^.]+)$`, billing.ChangePlanPreview),
			pipe5.POST(`^/billing/change-plan/(?P<new_plan_id>[^.]+)$`, billing.ChangePlan),
			pipe5.PathGET("/billing/upgrade", billing.Upgrade),
			pipe5.GET(`^/billing/upgrade/(?P<plain_id>[^.]+)$`, billing.UpgradeToPlan),
			pipe5.GET(`^/billing/upgrade/enterprise/(?P<plain_id>[^.]+)$`, billing.UpgradeEnterprisePlan),
			pipe5.GET(`^/billing/change-plan/enterprise/(?P<plain_id>[^.]+)$`, billing.ChangeEnterprisePlan),
			pipe5.PathGET("/billing/upgrade-success", billing.UpgradeSuccess),
			pipe5.PathGET("/billing/subscription/ping", billing.PingSubscription),
		),

		plug.PREFIX("/sites",
			pipe5.And(plug.RequireAccount).PathGET("/sites", site.Index),
			pipe5.And(plug.RequireAccount).PathGET("/sites/new", site.New),
			pipe5.And(plug.RequireAccount).PathPOST("/sites", site.CreateSite),
			sitePipe.GET(`^/sites/stats/(?P<domain>[^.]+)$`, site.Stats),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/make-public$`, site.MakePublic),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/make-private$`, site.MakePrivate),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/weekly-report/enable$`, site.EnableWeeklyReport),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/weekly-report/disable$`, site.DisableWeeklyReport),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/weekly-report/recipients$`, site.AddWeeklyReportRecipient),
			sitePipe.DELETE(`^/sites/(?P<website>[^.]+)/weekly-report/recipients/(?P<recipient>[^.]+)$`, site.RemoveWeeklyReportRecipient),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/monthly-report/enable$`, site.EnableMonthlyReport),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/monthly-report/disable$`, site.DisableMonthlyReport),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/monthly-report/recipients$`, site.AddMonthlyReportRecipient),
			sitePipe.DELETE(`^/sites/(?P<website>[^.]+)/monthly-report/recipients/(?P<recipient>[^.]+)$`, site.RemoveMonthlyReportRecipient),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/spike-notification/enable$`, site.EnableSpikeNotification),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/spike-notification/disable$`, site.DisableSpikeNotification),
			sitePipe.PUT(`^/sites/(?P<website>[^.]+)/spike-notification$`, site.UpdateSpikeNotification),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/spike-notification/recipients$`, site.AddSpikeNotificationRecipient),
			sitePipe.DELETE(`^/sites/(?P<website>[^.]+)/spike-notification/recipients/(?P<recipient>[^.]+)$`, site.RemoveSpikeNotificationRecipient),
			sitePipe.GET(`^/sites/(?P<website>[^.]+)/shared-links/new$`, site.NewSharedLink),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/shared-links$`, site.CreateSharedLink),
			sitePipe.GET(`^/sites/(?P<website>[^.]+)/shared-links/(?P<slug>[^.]+)/edit$`, site.EditSharedLink),
			sitePipe.PUT(`^/sites/(?P<website>[^.]+)/shared-links/(?P<slug>[^.]+)$`, site.UpdateSharedLink),
			sitePipe.DELETE(`^/sites/(?P<website>[^.]+)/shared-links/(?P<slug>[^.]+)$`, site.DeleteSharedLink),
			sitePipe.DELETE(`^/sites/(?P<website>[^.]+)/custom-domains/(?P<id>[^.]+)$`, site.DeleteCustomDomain),
			sitePipe.GET(`^/sites/(?P<website>[^.]+)/memberships/invite$`, site.InviteMemberForm),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/memberships/invite$`, site.InviteMember),
			sitePipe.POST(`^/sites/invitations/(?P<invitation_id>[^.]+)/accept$`, site.AcceptInvitation),
			sitePipe.POST(`^/sites/invitations/(?P<invitation_id>[^.]+)/reject$`, site.RejectInvitation),
			sitePipe.DELETE(`^/sites/(?P<website>[^.]+)/invitations/(?P<invitation_id>[^.]+)/reject$`, site.RemoveInvitation),
			sitePipe.GET(`^/sites/(?P<website>[^.]+)/transfer-ownership$`, site.TransferOwnershipForm),
			sitePipe.POST(`^/sites/(?P<website>[^.]+)/transfer-ownership$`, site.TransferOwnership),
			sitePipe.PUT(`^/sites/(?P<website>[^.]+)/memberships/(?P<id>[^.]+)/role/(?P<new_role>[^.]+)$`, site.UpdateRole),
			sitePipe.DELETE(`^/sites/(?P<website>[^.]+)/memberships/(?P<id>[^.]+)$`, site.RemoveMember),
			sitePipe.GET(`^/sites/(?P<website>[^.]+)/weekly-report/unsubscribe$`, site.WeeklyReport),
			sitePipe.GET(`^/sites/(?P<website>[^.]+)/monthly-report/unsubscribe$`, site.MonthlyReport),
			NotFound,
		),

		sitePipe.GET(`^/(?P<website>[^.]+)/snippet$`, site.AddSnippet),
		sitePipe.GET(`^/(?P<website>[^.]+)/settings$`, site.Settings),
		sitePipe.GET(`^/(?P<website>[^.]+)/settings/general$`, site.SettingsGeneral),
		sitePipe.GET(`^/(?P<website>[^.]+)/settings/people$`, site.SettingsPeople),
		sitePipe.GET(`^/(?P<website>[^.]+)/settings/visibility$`, site.SettingsVisibility),
		sitePipe.GET(`^/(?P<website>[^.]+)/settings/goals$`, site.SettingsGoals),
		sitePipe.GET(`^/(?P<website>[^.]+)/settings/search-console$`, site.SettingsSearchConsole),
		sitePipe.GET(`^/(?P<website>[^.]+)/settings/email-reports$`, site.SettingsEmailReports),
		sitePipe.GET(`^/(?P<website>[^.]+)/settings/custom-domain$`, site.SettingsCustomDomain),
		sitePipe.GET(`^/(?P<website>[^.]+)/settings/danger-zone$`, site.SettingsDangerZone),
		sitePipe.GET(`^/(?P<website>[^.]+)/goals/new$`, site.NewGoal),
		sitePipe.POST(`^/(?P<website>[^.]+)/goals$`, site.CreateGoal),
		sitePipe.DELETE(`^/(?P<website>[^.]+)/goals/(?P<id>[^.]+)$`, site.DeleteGoal),
		sitePipe.PUT(`^/(?P<website>[^.]+)/settings$`, site.UpdateSettings),
		sitePipe.PUT(`^/(?P<website>[^.]+)/settings/google$`, site.UpdateGoogleAuth),
		sitePipe.DELETE(`^/(?P<website>[^.]+)/settings/google-search$`, site.DeleteGoogleAuth),
		sitePipe.DELETE(`^/(?P<website>[^.]+)/settings/google-import$`, site.DeleteGoogleAuth),
		sitePipe.DELETE(`^/(?P<website>[^.]+)$`, site.DeleteSite),
		sitePipe.DELETE(`^/(?P<website>[^.]+)/stats$`, site.ResetStats),
		sitePipe.GET(`^/(?P<website>[^.]+)/import/google-analytics/view-id$`, site.ImportFromGoogleViewIdForm),
		sitePipe.POST(`^/(?P<website>[^.]+)/import/google-analytics/view-id$`, site.ImportFromGoogleViewId),
		sitePipe.GET(`^/(?P<website>[^.]+)/import/google-analytics/user-metric$`, site.ImportFromGoogleUserMetricNotice),
		sitePipe.GET(`^/(?P<website>[^.]+)/import/google-analytics/confirm$`, site.ImportFromGoogleConfirm),
		sitePipe.POST(`^/(?P<website>[^.]+)/settings/google-import$`, site.ImportFromGoogle),
		sitePipe.DELETE(`^/(?P<website>[^.]+)/settings/forget-imported$`, site.ForgetImported),
		sitePipe.GET(`^/(?P<domain>[^.]+)/export$`, site.CsvExport),
		NotFound,
	}

}

func NotFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.ERROR(r.Context(), w, http.StatusNotFound)
	})
}

func Load(cfg *config.Config) plug.Plug {
	load := cfg.Env == config.Config_load
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if load && r.Method == http.MethodPost && r.URL.Path == "/api/event" {
				api.Events(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
