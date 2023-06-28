package goals

import (
	"fmt"
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/internal/templates"
	"gorm.io/gorm"
)

var newGoalTpl = templates.Focus("site/new_goal.html")

func Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	site := models.GetSite(ctx)
	name := r.Form.Get("name")
	event := r.Form.Get("event_name")
	path := r.Form.Get("page_path")
	exists := models.Exists(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.Goal{}).Where("name = ?", name)
	})
	valid := models.ValidateGoals(event, path)
	if exists || !valid {
		r = sessions.SaveCsrf(w, r)
		render.HTML(ctx, w, newGoalTpl, http.StatusOK, func(ctx *templates.Context) {
			ctx.Site = site
			if !valid {
				ctx.Errors["event_name"] = "this field is required and cannot be blank"
				ctx.Errors["page_path"] = "this field is required and must start with a /"
			}
			if exists {
				ctx.Errors["name"] = "goal already exists"
			}
			ctx.Form = r.Form
		})
		return
	}
	models.CreateGoal(ctx, site, name, event, path)
	session, r := sessions.Load(r)
	session.Success("Goal created successfully").Save(ctx, w)
	to := fmt.Sprintf("/%s/%s/settings#goals", u.Name, site.Domain)
	http.Redirect(w, r, to, http.StatusFound)
}
