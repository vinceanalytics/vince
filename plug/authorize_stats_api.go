package plug

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func AuthorizeStatsAPI(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := bearer(r.Header)
		if token == "" {
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing API key. Please use a valid Vince API key as a Bearer Token."
			})
			return
		}
		hashedToken := models.HashAPIKey(ctx, token)
		key := models.KeyByHash(ctx, hashedToken)
		if key == nil {
			render.ERROR(ctx, w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing API key. Please use a valid Vince API key as a Bearer Token."
			})
			return
		}
		rate, burst := key.RateLimit()
		if !caches.AllowAPI(ctx, key.ID, rate, burst) {
			render.ERROR(r.Context(), w, http.StatusTooManyRequests, func(ctx *templates.Context) {
				ctx.Error.StatusText = fmt.Sprintf(
					"Too many API requests. Your API key is limited to %d requests per hour.",
					key.HourlyAPIRequestLimit,
				)
			})
			return
		}
		siteID := r.URL.Query().Get("site_id")
		if siteID == "" {
			render.ERROR(r.Context(), w, http.StatusBadRequest, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing site ID. Please provide the required site_id parameter with your request."
			})
			return
		}
		site := models.SiteByDomain(ctx, siteID)
		if site == nil {
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Invalid API key or site ID. Please make sure you're using a valid API key with access to the site you've requested."
			})
			return
		}
		isSuperUser := config.Get(r.Context()).IsSuperUser(key.UserID)
		isMember := site.IsMember(ctx, key.UserID)
		switch {
		case isSuperUser, isMember:
			r = r.WithContext(models.SetSite(ctx, site))
		case site.Locked:
			render.ERROR(r.Context(), w, http.StatusPaymentRequired, func(ctx *templates.Context) {
				ctx.Error.StatusText = "This Vince site is locked due to missing active subscription. In order to access it, the site owner should subscribe to a suitable plan"
			})
			return
		default:
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Invalid API key or site ID. Please make sure you're using a valid API key with access to the site you've requested."
			})
			return
		}
		h.ServeHTTP(w, r)
	})
}

func bearer(h http.Header) string {
	a := h.Get("authorization")
	if a == "" {
		return ""
	}
	if !strings.HasPrefix(a, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(a, "Bearer "))
}
