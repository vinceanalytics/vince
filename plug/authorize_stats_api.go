package plug

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/templates"
)

func AuthorizeStatsAPI(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tokenString := bearer(r.Header)
		if tokenString == "" {
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing API key. Please use a valid Vince API key as a Bearer Token."
			})
			return
		}
		claims := models.VerifyAPIKey(ctx, tokenString)
		if claims == nil {
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing API key. Please use a valid Vince API key as a Bearer Token."
			})
			return
		}
		key := models.APIKeyByID(ctx, claims.ID)
		if key == nil {
			render.ERROR(ctx, w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing API key. Please use a valid Vince API key as a Bearer Token."
			})
			return
		}
		rate, burst := models.APIRateLimit(key)
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
		isMember := models.UserIsMember(ctx, key.UserID, site.ID)
		switch {
		case isSuperUser, isMember:
			r = r.WithContext(models.SetSite(ctx, site))
		default:
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Invalid API key or site ID. Please make sure you're using a valid API key with access to the site you've requested."
			})
			return
		}
		models.UpdateAPIKeyUse(ctx, claims.ID)
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
