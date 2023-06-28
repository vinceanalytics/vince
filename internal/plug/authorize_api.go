package plug

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/params"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/pkg/schema"
)

func AuthAPI(resource schema.Resource, action schema.Verb) Plug {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := bearer(r.Header)
			if tokenString == "" {
				render.JSONError(w, http.StatusUnauthorized,
					"Missing API key. Please use a valid vince API key as a Bearer Token.",
				)
				return
			}
			ctx := r.Context()
			params := params.Get(ctx)
			owner := params.Get("owner")
			site := params.Get("site")
			claims := models.GetApiKey(ctx, tokenString)
			if claims == nil || !claims.Can(ctx, owner, site, resource, action) {
				render.JSONError(w, http.StatusUnauthorized,
					"Invalid API key. Please make sure you're using a valid API key with access to the resource you've requested.",
				)
				return
			}

			user := models.QueryUserByNameOrEmail(ctx, claims.Owner)
			if user == nil {
				render.JSONError(w, http.StatusUnauthorized,
					"Invalid API key. Please make sure you're using a valid API key with access to the resource you've requested.",
				)
				return
			}
			ctx = models.SetUser(ctx, user)
			if site != "" {
				s := models.SiteByDomain(ctx, site)
				if s == nil {
					render.JSONError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
					return
				}
				ctx = models.SetSite(ctx, s)
			}
			r = r.WithContext(ctx)

			models.UpdatePersonalAccessTokenUse(ctx, claims.ID)
			h.ServeHTTP(w, r)
		})

	}
}
