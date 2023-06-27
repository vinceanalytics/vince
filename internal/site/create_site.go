package site

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/internal/templates"
)

func Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	domain := r.Form.Get("domain")
	domain, bad := models.ValidateSiteDomain(ctx, domain)
	if bad != "" {
		r = sessions.SaveCsrf(w, r)
		render.HTML(r.Context(), w, templates.SiteNew, http.StatusOK, func(ctx *templates.Context) {
			if bad != "" {
				ctx.Errors["domain"] = bad
			}
			ctx.Form = r.Form
		})
		return
	}
	if !models.CreateSite(ctx, u, domain, r.Form.Get("public") == "true") {
		render.ERROR(r.Context(), w, http.StatusInternalServerError)
		return
	}
	ss, r := sessions.Load(r)
	ss.Save(ctx, w)
	to := fmt.Sprintf("/%s/%s", u.Name, url.PathEscape(domain))
	http.Redirect(w, r, to, http.StatusFound)
}
