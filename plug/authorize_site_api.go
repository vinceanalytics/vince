package plug

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func AuthorizeSiteAPI(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearer(r.Header)
		if token == "" {
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.StatusText = "Missing API key. Please use a valid vince API key as a Bearer Token."
			})
			return
		}
		hashedToken := models.HashAPIKey(r.Context(), token)
		key := new(models.APIKey)
		err := models.Get(r.Context()).Where("key_hash=?", hashedToken).
			Where("scopes like ?", "sites:provision:*").First(key).Error
		if err != nil {
			log.Get(r.Context()).Err(err).Msg("failed to get hashed api key")
			render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
				ctx.StatusText = "Missing API key. Please use a valid vince API key as a Bearer Token."
			})
			return
		}
		user := new(models.User)
		err = models.Get(r.Context()).First(user, key.UserID).Error
		if err != nil {
			log.Get(r.Context()).Err(err).
				Uint64("user_id", key.UserID).Msg("failed to get user from api key")
			render.ERROR(r.Context(), w, http.StatusInternalServerError)
			return
		}
		r = r.WithContext(models.SetCurrentUser(r.Context(), user))
		h.ServeHTTP(w, r)
	})
}
