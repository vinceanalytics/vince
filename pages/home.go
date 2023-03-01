package pages

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/render"
)

func Home(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, templates.Home, http.StatusOK)
}
