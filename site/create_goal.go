package site

import (
	"fmt"
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
)

func CreateGoal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	event := r.Form.Get("event_name")
	path := r.Form.Get("page_path")
	if !models.ValidateGoals(event, path) {
		r = sessions.SaveCsrf(w, r)
		render.HTML(ctx, w, templates.SiteNewGoal, http.StatusOK, func(ctx *templates.Context) {
			ctx.Site = site
			ctx.Errors["event_name"] = "this field is required and cannot be blank"
			ctx.Errors["page_path"] = "this field is required and must start with a /"
			ctx.Form = r.Form
		})
		return
	}
	models.CreateGoal(ctx, site.Domain, event, path)
	session, r := sessions.Load(r)
	session.SuccessFlash("Goal created successfully").Save(w)
	to := fmt.Sprintf("/%s/settings/goals", site.SafeDomain())
	http.Redirect(w, r, to, http.StatusFound)
}
