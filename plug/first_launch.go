package plug

import (
	"net/http"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/models"
)

func FirstLaunch(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.Get(r.Context()).IsSelfHost {
			var count int64
			models.Get(r.Context()).Model(&models.User{}).Count(&count)
			if count == 0 {
				http.Redirect(w, r, "/register", http.StatusFound)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
