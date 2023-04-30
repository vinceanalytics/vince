package api

import (
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/params"
	"github.com/gernest/vince/render"
)

func DomainStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.SiteByDomain(ctx, params.Get(ctx)["domain"])
	if site != nil {
		status := "waiting"
		if site.HasStats {
			status = "ready"
		}
		render.JSON(w, http.StatusOK, status)
		return
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
