package plug

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

func AuthorizeSiteAPI(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := bearer(r.Header)
		if tokenString == "" {
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing API key. Please use a valid vince API key as a Bearer Token."
			})
			return
		}
		ctx := r.Context()
		claims := models.VerifyAPIKey(ctx, tokenString)
		if claims == nil {
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing API key. Please use a valid vince API key as a Bearer Token."
			})
			return
		}
		user := models.UserByID(ctx, claims.UserID)
		if user == nil {
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.Error.StatusText = "Missing API key. Please use a valid vince API key as a Bearer Token."
			})
			return
		}
		r = r.WithContext(models.SetUser(r.Context(), user))
		models.UpdateAPIKeyUse(ctx, claims.ID)
		h.ServeHTTP(w, r)
	})
}
