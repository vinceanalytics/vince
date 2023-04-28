package site

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
)

func Stats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	role := models.GetRole(ctx)
	canSeeStats := !site.Locked || role == "super_admin"
	switch {
	case !site.StatsStartDate.IsZero() && canSeeStats:
		w.Header().Set("x-robots-tag", "noindex")
		var offer bool
		session, _ := sessions.Load(r)
		if session.Data.EmailReport != nil {
			offer = session.Data.EmailReport[site.Domain]
		}
		hasGoals := models.SiteHasGoals(ctx, site.Domain)
		render.HTML(ctx, w, templates.Stats, http.StatusOK, func(ctx *templates.Context) {
			ctx.Site = site
			ctx.Title = "Vince Analytics  · " + site.Domain
			ctx.EmailReport = offer
			ctx.HasGoals = hasGoals
		})
	case site.StatsStartDate.IsZero() && canSeeStats:
		render.HTML(ctx, w, templates.WaitingFirstPageView, http.StatusOK, func(ctx *templates.Context) {
			ctx.Site = site
		})
	case site.Locked:
		owner := models.SiteOwner(ctx, site.ID)
		render.HTML(ctx, w, templates.SiteLocked, http.StatusOK, func(ctx *templates.Context) {
			ctx.Site = site
			ctx.Owner = owner
		})
	default:
		render.ERROR(r.Context(), w, http.StatusInternalServerError)
	}
}
