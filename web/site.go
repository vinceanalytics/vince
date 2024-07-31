package web

import (
	"net/http"

	"github.com/gernest/len64/web/db"
	"github.com/gernest/len64/web/db/schema"
)

func CreateSiteForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	createSite.Execute(w, db.Context(make(map[string]any)))
}

func CreateSite(db *db.Config, w http.ResponseWriter, r *http.Request) {
	createSite.Execute(w, db.Context(make(map[string]any)))
}

func Sites(db *db.Config, w http.ResponseWriter, r *http.Request) {
	usr := db.CurrentUser()
	usr.Preload(db.Get(), "Sites", "Invitations")

	invites := make([]map[string]any, 0, len(usr.Invitations))
	sites := make([]map[string]any, 0, len(usr.Invitations))
	for _, i := range usr.Invitations {
		m := map[string]any{
			"id":   i.ID,
			"role": i.Role,
		}
		site := new(schema.Site)
		site.ByID(db.Get(), i.SiteID)
		m["site"] = map[string]any{
			"id":     site.ID,
			"domain": site.Domain,
		}
		m["visitors"] = 0
		invites = append(invites, m)
	}

	for _, s := range usr.Sites {
		sites = append(sites, map[string]any{
			"id":       s.ID,
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
