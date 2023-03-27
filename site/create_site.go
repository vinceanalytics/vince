package site

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
	"github.com/miekg/dns"
)

func CreateSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetCurrentUser(ctx)
	owned := u.CountOwnedSites(ctx)
	domain := r.Form.Get("domain")
	limit := u.SitesLimit(ctx)
	if _, ok := dns.IsDomainName(domain); !ok {
		sessions.SaveCsrf(w, r)
		render.HTML(r.Context(), w, templates.SiteNew, http.StatusOK, func(ctx *templates.Context) {
			ctx.Errors = map[string]string{
				"domain": "not a valid domain name",
			}
			ctx.NewSite = &templates.NewSite{
				IsFirstSite: owned == 0,
				IsAtLimit:   owned >= int64(limit),
				SiteLimit:   limit,
			}
		})
		return
	}
	domain = dns.CanonicalName(domain)
	err := models.Get(ctx).Model(u).Association("Sites").Append(&models.Site{
		Domain: domain,
		Public: r.Form.Get("public") == "true",
	})
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed saving site")
		render.ERROR(r.Context(), w, http.StatusInternalServerError)
		return
	}
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
