package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/timex"
)

func Stats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	role := models.GetRole(ctx)
	owner := models.SiteOwner(ctx, site.ID)
	var canSeeStats bool
	switch role {
	case "super_admin":
	default:
		canSeeStats = true
	}
	if canSeeStats {
		w.Header().Set("x-robots-tag", "noindex")
		var offer bool
		session, _ := sessions.Load(r)
		if session.Data.EmailReport != nil {
			offer = session.Data.EmailReport[site.Domain]
		}
		hasGoals := models.SiteHasGoals(ctx, site.Domain)
		q := r.URL.Query()
		period := timex.Parse(q.Get("o"))
		key := q.Get("k")
		if key == "" {
			key = timeseries.BaseKey
		}
		prop := timeseries.ParseProperty(q.Get("p"))
		start := core.Now(ctx)
		window := period.Window(start)
		stats := timeseries.Root(ctx, owner.ID, site.ID, timeseries.RootOptions{
			Start:  start,
			Window: window,
			Key:    key,
		})
		stats.Period = period
		stats.Domain = site.Domain
		stats.Key = key
		stats.Prop = prop
		render.HTML(ctx, w, templates.Stats, http.StatusOK, func(ctx *templates.Context) {
			ctx.Site = site
			ctx.Title = "Vince Analytics  · " + site.Domain
			ctx.EmailReport = offer
			ctx.HasGoals = hasGoals
			ctx.Stats = &stats
		})
		return
	}
	render.ERROR(ctx, w, http.StatusUnauthorized)
}
