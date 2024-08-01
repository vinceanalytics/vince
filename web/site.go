package web

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gernest/len64/internal/kv"
	"github.com/gernest/len64/web/db"
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
	_, err := usr.CreateSite(db.Get(), domain, r.Form.Get("public") == "true")
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
