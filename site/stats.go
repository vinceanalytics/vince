package site

import (
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

func Stats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	role := models.GetRole(ctx)
	canSeeStats := !site.Locked || role == "super_admin"
	switch {
	case !site.StatsStartDate.IsZero() && canSeeStats:
	case site.StatsStartDate.IsZero() && canSeeStats:
	case site.Locked:
	default:
		render.ERROR(r.Context(), w, http.StatusNotImplemented)
	}
}
