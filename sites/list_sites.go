package sites

import (
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func ListSites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := models.GetUser(ctx)
	models.PreloadUser(ctx, user, "Sites")
	render.JSON(w, http.StatusOK, user.Sites)
}
