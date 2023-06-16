package site

import (
	"net/http"
	"time"

	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/property"
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
		period := timex.Parse(q.Get("w"))
		key := q.Get("k")
		if key == "" {
			key = timeseries.BaseKey
		}
		prop := property.ParseProperty(q.Get("p"))
		start := core.Now(ctx)
		window := period.Window(start)
		var stats timeseries.Stats
		stats.Result = timeseries.Query(ctx, owner.ID, site.ID,
			buildQuery(window, prop, key),
		)
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

func buildQuery(
	window time.Duration,
	prop property.Property,
	key string,
) query.Query {
	m := &query.Metrics{
		Visitors:       &query.Select{Exact: &query.Value{Value: key}},
		Views:          &query.Select{Exact: &query.Value{Value: key}},
		Events:         &query.Select{Exact: &query.Value{Value: key}},
		Visits:         &query.Select{Exact: &query.Value{Value: key}},
		BounceRates:    &query.Select{Exact: &query.Value{Value: key}},
		VisitDurations: &query.Select{Exact: &query.Value{Value: key}},
	}
	if prop != property.Base {
		p := &query.Props{}
		p.Set(prop.String(), m)
		return query.Query{
			Window: &query.Duration{Value: window},
			Props:  p,
		}
	}
	all := &query.Metrics{
		Visitors:       &query.Select{},
		Views:          &query.Select{},
		Events:         &query.Select{},
		Visits:         &query.Select{},
		BounceRates:    &query.Select{},
		VisitDurations: &query.Select{},
	}
	p := &query.Props{}
	p.All(func(s string, mq *query.Metrics) {
		if s == property.Base.String() {
			// Optimize this by doing exact match.
			*mq = *m
		} else {
			*mq = *all
		}
	})
	return query.Query{
		Window: &query.Duration{Value: window},
		Props:  p,
	}
}
