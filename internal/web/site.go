package web

import (
	"fmt"
	"net/http"
	"net/url"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/db/plug"
)

func CreateSiteForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, createSite, nil)
}

func CreateSite(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	domain := r.Form.Get("domain")
	domain, bad := db.Get().ValidateSiteDomain(domain)
	if bad != "" {
		db.SaveCsrf(w)
		createSite.Execute(w, db.Context(map[string]any{
			"validation_domain": bad,
		}))
		return
	}
	err := db.Get().CreateSite(domain, false)
	if err != nil {
		db.HTMLCode(http.StatusInternalServerError, w, e500, nil)
		db.Logger().Error("creating site", "err", err)
		return
	}
	to := fmt.Sprintf("/%s/snippet", url.PathEscape(domain))
	http.Redirect(w, r, to, http.StatusFound)
}

func Sites(db *db.Config, w http.ResponseWriter, r *http.Request) {
	sites := make([]map[string]any, 0, 16)
	db.Get().Domains(func(s *v1.Site) {
		sites = append(sites, map[string]any{
			"domain":   s.Domain,
			"public":   s.Public,
			"locked":   s.Locked,
			"visitors": 0,
		})
	})
	ctx := make(map[string]any)

	if len(sites) > 0 {
		ctx["sites"] = sites
	}
	db.HTML(w, sitesIndex, ctx)
}

func AddSnippet(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, addSnippet, nil)
}

func Unimplemented(db *db.Config, w http.ResponseWriter, r *http.Request) {
}

func RequireSiteAccess(h plug.Handler) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		domain := r.PathValue("domain")
		site := db.Get().Site(domain)
		if site == nil {
			db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
			return
		}
		if db.CurrentUser() != nil {
			db.SetSite(site)
			h(db, w, r)
			return
		}
		if !site.Public {
			db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
			return
		}
		db.SetSite(site)
		h(db, w, r)
	}
}
