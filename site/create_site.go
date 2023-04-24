package site

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
)

func CreateSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	owned := u.CountOwnedSites(ctx)
	domain := r.Form.Get("domain")
	limit := u.SitesLimit(ctx)
	isAtLimit := limit != -1 && owned >= int64(limit)
	domain, bad := models.ValidateSiteDomain(ctx, domain)
	if isAtLimit || bad != "" {
		r = sessions.SaveCsrf(w, r)
		render.HTML(r.Context(), w, templates.SiteNew, http.StatusOK, func(ctx *templates.Context) {
			if bad != "" {
				ctx.Errors["domain"] = bad
			}
			ctx.Form = r.Form
			ctx.NewSite = &templates.NewSite{
				IsFirstSite: owned == 0,
				IsAtLimit:   isAtLimit,
				SiteLimit:   limit,
			}
		})
		return
	}
	if !models.CreateSite(ctx, u, domain, r.Form.Get("public") == "true") {
		render.ERROR(r.Context(), w, http.StatusInternalServerError)
		return
	}
	ss, r := sessions.Load(r)
	if ss.Data.EmailReport == nil {
		ss.Data.EmailReport = make(map[string]bool)
	}
	ss.Data.EmailReport[domain] = true
	ss.Save(w)
	to := fmt.Sprintf("/%s/snippet", url.PathEscape(domain))
	http.Redirect(w, r, to, http.StatusFound)
}
