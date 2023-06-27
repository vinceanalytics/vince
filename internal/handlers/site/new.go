package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

func New(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, templates.SiteNew, http.StatusOK, func(ctx *templates.Context) {
	})
}
