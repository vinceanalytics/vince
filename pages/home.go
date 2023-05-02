package pages

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func Home(w http.ResponseWriter, r *http.Request) {
	cts := r.Context()
	if models.GetUser(cts) != nil {
		http.Redirect(w, r, "/sites", http.StatusFound)
		return
	}
	render.HTML(r.Context(), w, templates.Home, http.StatusOK)
}
