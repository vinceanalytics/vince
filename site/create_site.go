package site

import (
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
	"github.com/miekg/dns"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func CreateSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	owned := u.CountOwnedSites(ctx)
	domain := r.Form.Get("domain")
	limit := u.SitesLimit(ctx)
	isAtLimit := limit != -1 && owned >= int64(limit)
	_, valid := dns.IsDomainName(domain)
	var exists bool
	if !isAtLimit && valid {
		exists = models.Exists(ctx, func(db *gorm.DB) *gorm.DB {
			return db.Model(&models.Site{}).
				Joins("left join  site_memberships on sites.id = site_memberships.site_id and site_memberships.role = 'owner' and  site_memberships.user_id = ? and sites.domain = ?", u.ID, domain)
		})
	}
	if isAtLimit || !valid || exists {
		sessions.SaveCsrf(w, r)
		render.HTML(r.Context(), w, templates.SiteNew, http.StatusOK, func(ctx *templates.Context) {
			if !valid {
				ctx.Errors = map[string]string{
					"domain": "not a valid domain name",
				}
			}
			if exists {
				ctx.Errors = map[string]string{
					"domain": "already exists",
				}
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
	db := models.Get(ctx)
	db.Logger = db.Logger.LogMode(logger.Info)
	err := models.Get(ctx).Model(u).Association("Sites").Append(&models.Site{
		Domain: domain,
		Public: r.Form.Get("public") == "true",
	})
	if err != nil {
		log.Get(ctx).Err(err).Msg("failed saving site")
		render.ERROR(r.Context(), w, http.StatusInternalServerError)
		return
	}
	ss, r := sessions.Load(r)
	if ss.Data.EmailReport == nil {
		ss.Data.EmailReport = map[string]bool{
			domain: true,
		}
	} else {
		ss.Data.EmailReport[domain] = true
	}
	ss.Save(w)
	path := "/" + domain + "/snippet"
	http.Redirect(w, r, path, http.StatusFound)
}
