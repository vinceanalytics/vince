package router

import (
	"context"
	"net/http"

	"github.com/gernest/vince/api"
	"github.com/gernest/vince/auth"
	"github.com/gernest/vince/plug"
	"github.com/gernest/vince/render"
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

		pipe2.GET(`^/share/(?P<domain>[^.]+)$`, S501),
		pipe2.GET(`^/share/(?P<slug>[^.]+)/authenticate$`, S501),

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
		pipe5.PathGET("/password", S501),
		pipe5.PathPOST("/password", S501),
		pipe5.PathGET("/logout", S501),
		pipe5.PathGET("/settings", S501),
		pipe5.PathPUT("/settings", S501),
		pipe5.PathGET("/me", S501),
		pipe5.PathGET("/settings/api-keys/new", S501),
		pipe5.PathPOST("/settings/api-keys/new", S501),
		pipe5.PathGET("/auth/google/callback", S501),
		pipe5.PathGET("/billing/change-plan", S501),
		pipe5.PathGET("/billing/upgrade", S501),
		pipe5.PathGET("/billing/subscription/ping", S501),
		pipe5.PathGET("sites", S501),
		pipe5.PathPOST("/sites", S501),
		pipe5.PathGET("/sites/new", S501),
		pipe5.GET(`^/register/invitation/(?P<invitation_id>[^.]+)$`, S501),
		pipe5.POST(`^/register/invitation/(?P<invitation_id>[^.]+)$`, S501),
		pipe5.DELETE(`^/settings/api-keys/(?P<id>[^.]+)$`, S501),
		pipe5.GET(`^/billing/change-plan/preview/(?P<plan_id>[^.]+)$`, S501),
		pipe5.POST(`^/billing/change-plan/(?P<new_plan_id>[^.]+)$`, S501),
		pipe5.GET(`^/billing/upgrade/(?P<plain_id>[^.]+)$`, S501),
		pipe5.GET(`^/billing/upgrade/enterprise/(?P<plain_id>[^.]+)$`, S501),
		pipe5.GET(`^/billing/change-plan/enterprise/(?P<plain_id>[^.]+)$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/make-public$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/make-private$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/weekly-report/enable$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/weekly-report/disable$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/weekly-report/recipients$`, S501),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/weekly-report/recipients/(?P<recipient>[^.]+)$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/monthly-report/enable$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/monthly-report/disable$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/monthly-report/recipients$`, S501),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/monthly-report/recipients/(?P<recipient>[^.]+)$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/spike-notification/enable$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/spike-notification/disable$`, S501),
		pipe5.PUT(`^/sites/(?P<website>[^.]+)/spike-notification$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/spike-notification/recipients$`, S501),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/spike-notification/recipients/(?P<recipient>[^.]+)$`, S501),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/shared-links/new$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/shared-links$`, S501),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/shared-links/(?P<slug>[^.]+)/edit$`, S501),
		pipe5.PUT(`^/sites/(?P<website>[^.]+)/shared-links/(?P<slug>[^.]+)$`, S501),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/shared-links/(?P<slug>[^.]+)$`, S501),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/custom-domains/(?P<id>[^.]+)$`, S501),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/memberships/invite$`, S501),
		pipe5.POST(`^/sites/(?P<website>[^.]+)/memberships/invite$`, S501),
		pipe5.POST(`^/sites/invitations/(?P<invitation_id>[^.]+)/accept$`, S501),
		pipe5.POST(`^/sites/invitations/(?P<invitation_id>[^.]+)/reject$`, S501),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/transfer-ownership$`, S501),
		pipe5.PUT(`^/sites/(?P<website>[^.]+)/memberships/(?P<id>[^.]+)/role/(?P<new_role>[^.]+)$`, S501),
		pipe5.DELETE(`^/sites/(?P<website>[^.]+)/memberships/(?P<id>[^.]+)$`, S501),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/weekly-report/unsubscribe$`, S501),
		pipe5.GET(`^/sites/(?P<website>[^.]+)/monthly-report/unsubscribe$`, S501),

		pipe5.GET(`^/(?P<website>[^.]+)/snippet$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/settings$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/general$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/people$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/visibility$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/goals$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/search-console$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/email-reports$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/custom-domain$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/settings/danger-zone$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/goals/new$`, S501),
		pipe5.POST(`^/(?P<website>[^.]+)/goals$`, S501),
		pipe5.DELETE(`^/(?P<website>[^.]+)/goals/(?P<id>[^.]+)$`, S501),
		pipe5.PUT(`^/(?P<website>[^.]+)/settings$`, S501),
		pipe5.PUT(`^/(?P<website>[^.]+)/settings/google$`, S501),
		pipe5.DELETE(`^/(?P<website>[^.]+)/settings/google-search$`, S501),
		pipe5.DELETE(`^/(?P<website>[^.]+)/settings/google-import$`, S501),
		pipe5.DELETE(`^/(?P<website>[^.]+)$`, S501),
		pipe5.DELETE(`^/(?P<website>[^.]+)/stats$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/import/google-analytics/view-id$`, S501),
		pipe5.POST(`^/(?P<website>[^.]+)/import/google-analytics/view-id$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/import/google-analytics/user-metric$`, S501),
		pipe5.GET(`^/(?P<website>[^.]+)/import/google-analytics/confirm$`, S501),
		pipe5.POST(`^/(?P<website>[^.]+)/settings/google-import$`, S501),
		pipe5.DELETE(`^/(?P<website>[^.]+)/settings/forget-imported$`, S501),
		pipe5.GET(`^/(?P<domain>[^.]+)/export$`, S501),
		pipe5.GET(`^/(?P<domain>[^.]+)/(?P<path>[^.]+)$`, S501),
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
