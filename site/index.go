package site

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/render"
)

func Index(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, templates.Sites, http.StatusOK, func(ctx *templates.Context) {
		ctx.Page = "sites"
	})
}
