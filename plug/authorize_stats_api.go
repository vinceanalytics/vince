package plug

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/limit"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func AuthorizeStatsAPI(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearer(r.Header)
		if token == "" {
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing API key. Please use a valid Vince API key as a Bearer Token."
			})
			return
		}
		hashedToken := models.HashAPIKey(r.Context(), token)
		key := new(models.APIKey)
		err := models.Get(r.Context()).Where("key_hash=?", hashedToken).First(key).Error
		if err != nil {
			log.Get(r.Context()).Err(err).Msg("failed to get hashed api key")
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing API key. Please use a valid Vince API key as a Bearer Token."
			})
			return
		}
		if !limit.API.Allow(key.RateLimit()) {
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
		site := new(models.Site)
		err = models.Get(r.Context()).Where("domain=?", siteID).
			Preload("SiteMemberships", "user_id=?", key.UserID).First(site).Error
		if err != nil {
			log.Get(r.Context()).Err(err).Str("domain", siteID).Msg("failed to get site")
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Invalid API key or site ID. Please make sure you're using a valid API key with access to the site you've requested."
			})
			return
		}
		switch {
		case config.Get(r.Context()).IsSuperUser(key.UserID),
			site.IsMember(key.UserID):
			r = r.WithContext(models.SetSite(r.Context(), site))
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
