package api

import (
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func Sites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr := models.GetUser(ctx)
	if usr != nil {
		models.PreloadUser(ctx, usr, "Sites")
		render.JSON(w, http.StatusOK, usr.Sites)
		return
	}
	render.ERROR(r.Context(), w, http.StatusUnauthorized)
}
