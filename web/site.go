package web

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gernest/len64/internal/kv"
	"github.com/gernest/len64/web/db"
	"github.com/gernest/len64/web/db/plug"
)

func CreateSiteForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	createSite.Execute(w, db.Context(make(map[string]any)))
}

func CreateSite(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	usr := db.CurrentUser()
	domain := r.Form.Get("domain")
	domain, bad := kv.ValidateSiteDomain(db.Get(), domain)
	if bad != "" {
		db.SaveCsrf(w)
		createSite.Execute(w, db.Context(map[string]any{
			"validation_domain": bad,
		}))
		return
	}
	// All sites are private by default
	_, err := usr.CreateSite(db.Get(), domain, false)
	if err != nil {
		db.HTMLCode(http.StatusInternalServerError, w, e500, nil)
		db.Logger().Error("creating site", "err", err)
		return
	}
	to := fmt.Sprintf("/%s/snippet", url.PathEscape(domain))
	http.Redirect(w, r, to, http.StatusFound)
}

func Sites(db *db.Config, w http.ResponseWriter, r *http.Request) {
	usr := db.CurrentUser()

	invites := make([]map[string]any, 0, len(usr.Invitations))
	sites := make([]map[string]any, 0, len(usr.Invitations))
	for _, i := range usr.Invitations {
		m := map[string]any{
			"id":   kv.FormatID(i.Id),
			"role": i.Role,
		}
		m["site"] = map[string]any{
			"id":     kv.FormatID(i.Site.Id),
			"domain": i.Site.Domain,
		}
		m["visitors"] = 0
		invites = append(invites, m)
	}

	for _, s := range usr.Sites {
		sites = append(sites, map[string]any{
			"id":       kv.FormatID(s.Id),
			"domain":   s.Domain,
			"visitors": 0,
		})
	}
	ctx := make(map[string]any)
	if len(invites) > 0 {
		ctx["invitations"] = invites
	}
	if len(sites) > 0 {
		ctx["sites"] = sites
	}
	db.HTML(w, sitesIndex, ctx)
}

func RequireSiteAccess(allowed ...string) plug.Middleware {
	if allowed == nil {
		allowed = []string{
			"public", "viewer", "admin", "super_admin", "owner",
		}
	}
	a := make(map[string]struct{}, len(allowed))
	for i := range allowed {
		a[allowed[i]] = struct{}{}
	}
	return func(h plug.Handler) plug.Handler {
		return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
			domain := r.PathValue("domain")

			if usr := db.CurrentUser(); usr != nil && usr.Site(domain) != nil {
				// Fast path the current user has some role with the asked domain
				site := usr.Site(domain)
				_, ok := a[site.Role.String()]
				if !ok {
					db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
					return
				}
				db.SetSite(site)
				h(db, w, r)
				return
			}
			usr := new(kv.User)
			usr.ByDomain(domain, db.Get())
			site := usr.Site(domain)
			if site == nil {
				db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
				return
			}
			role := ""
			switch {
			case db.CurrentUser() != nil && db.CurrentUser().SuperUser:
				role = "super_admin"
			case site.Public:
				role = "public"
			}
			_, accept := a[role]
			if accept {
				db.SetSite(site)
				h(db, w, r)
				return
			}
			db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
		}
	}
}
