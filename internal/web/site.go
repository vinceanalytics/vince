package web

import (
	"cmp"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/api/visitors"
	"github.com/vinceanalytics/vince/internal/domains"
	"github.com/vinceanalytics/vince/internal/web/db"
	"github.com/vinceanalytics/vince/internal/web/db/plug"
)

func NewGoalForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, newGoal, nil)
}

func GoalSettings(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	to := fmt.Sprintf("/%s/settings#goals", url.PathEscape(site.Domain))
	http.Redirect(w, r, to, http.StatusFound)
}

func CreateGoal(db *db.Config, w http.ResponseWriter, r *http.Request) {
	g := &v1.Goal{
		Name: r.FormValue("event_name"),
		Path: r.FormValue("page_path"),
	}
	if g.Name == "" && g.Path == "" {
		db.SaveCsrf(w)
		msg := "event_name or page_path must be set"
		db.HTML(w, newGoal, map[string]any{
			"error_event_name": msg,
			"error_page_path":  msg,
		})
		return
	}
	site := db.CurrentSite()
	site.Goals = append(site.Goals, g)
	db.Ops().Save(site)
	db.Success("Goal was successfully created")
	db.SaveSession(w)
	to := fmt.Sprintf("/%s/settings#goals", url.PathEscape(site.Domain))
	http.Redirect(w, r, to, http.StatusFound)
}

func DeleteGoal(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	idx, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err == nil {
		site.Goals = slices.Delete(site.Goals, idx, idx+1)
	}
	db.Ops().Save(site)
	db.Success("Goal was successfully deleted")
	db.SaveSession(w)
	to := fmt.Sprintf("/%s/settings#goals", url.PathEscape(site.Domain))
	http.Redirect(w, r, to, http.StatusFound)
}

func Status(db *db.Config, w http.ResponseWriter, r *http.Request) {
	body := "WAITING"
	if db.Ops().SeenFirstStats(db.CurrentSite().Domain) {
		body = "READY"
	}
	db.JSON(w, body)
}

func EditSharedLinksForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	db.HTML(w, edit, map[string]any{"slug": slug})
}

func SharedLinksForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, shared, nil)
}

func DeleteSharedLink(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	slug := r.PathValue("slug")
	err := db.Ops().DeleteSharedLink(site, slug)
	if err != nil {
		slog.Error("deleting shared link", "slug", slug, "domain", db.CurrentSite().Domain, "err", "err")
	}
	http.Redirect(w, r, fmt.Sprintf("/%s/settings#visibility", url.PathEscape(site.Domain)), http.StatusFound)
}

func EditSharedLink(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.Form.Get("name")
	site := db.CurrentSite()
	slug := r.PathValue("slug")
	err := db.Ops().EditSharedLink(site, slug, name)
	if err != nil {
		slog.Error("updating shared link", "slug", slug, "domain", db.CurrentSite().Domain, "err", "err")
	}
	http.Redirect(w, r, fmt.Sprintf("/%s/settings#visibility", url.PathEscape(site.Domain)), http.StatusFound)
}

func CreateSharedLink(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.Form.Get("name")
	password := r.Form.Get("password")
	site := db.CurrentSite()
	site.Public = true
	err := db.Ops().FindOrCreateCreateSharedLink(site.Domain, name, password)
	if err == nil {
		db.Logger().Error("failed creating shared link", "domain", db.CurrentSite().Domain)
	}
	http.Redirect(w, r, fmt.Sprintf("/%s/settings#visibility", url.PathEscape(site.Domain)), http.StatusFound)
}

func Settings(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, siteSettings, nil)
}

func Delete(db *db.Config, w http.ResponseWriter, r *http.Request) {
	domain := db.CurrentSite().Domain
	err := db.Ops().DeleteDomain(domain)
	if err != nil {
		slog.Error("deleting site", "domain", domain, "err", "err")
	}
	domains.Reload(db.Ops().Domains)
	http.Redirect(w, r, "/sites", http.StatusFound)
}

func MakePublic(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	site.Public = true
	err := db.Ops().Save(site)
	if err != nil {
		db.Logger().Error("making site  public", "domain", db.CurrentSite().Domain, "err", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/%s/settings#visibility", url.PathEscape(site.Domain)), http.StatusFound)
}

func MakePrivate(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	site.Public = false
	err := db.Ops().Save(site)
	if err != nil {
		db.Logger().Error("making site  private", "domain", db.CurrentSite().Domain, "err", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/%s/settings#visibility", url.PathEscape(site.Domain)), http.StatusFound)
}

func CreateSiteForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, createSite, nil)
}

func CreateSite(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	domain := r.Form.Get("domain")
	domain, bad := db.Ops().ValidateSiteDomain(domain)
	if bad != "" {
		db.SaveCsrf(w)
		createSite.Execute(w, db.Context(map[string]any{
			"validation_domain": bad,
		}))
		return
	}
	err := db.Ops().CreateSite(domain, false)
	if err != nil {
		db.HTMLCode(http.StatusInternalServerError, w, e500, nil)
		db.Logger().Error("creating site", "err", err)
		return
	}
	domains.Reload(db.Ops().Domains)
	to := fmt.Sprintf("/%s/snippet", url.PathEscape(domain))
	http.Redirect(w, r, to, http.StatusFound)
}

func SitesIndex(db *db.Config, w http.ResponseWriter, r *http.Request) {
	sites := make([]map[string]any, 0, 16)
	db.Ops().Domains(func(s *v1.Site) {
		sites = append(sites, map[string]any{
			"domain": s.Domain,
			"public": s.Public,
			"locked": s.Locked,
		})
	})
	db.JSON(w, map[string]any{
		"data": sites,
	})
}
func Sites(db *db.Config, w http.ResponseWriter, r *http.Request) {
	sites := make([]map[string]any, 0, 16)
	db.Ops().Domains(func(s *v1.Site) {
		sites = append(sites, map[string]any{
			"domain": s.Domain,
			"public": s.Public,
			"locked": s.Locked,
		})
	})

	for i := range sites {
		// compute visitors
		dom := sites[i]["domain"].(string)
		vs, err := visitors.Visitors(r.Context(), db.TimeSeries(), dom)
		if err != nil {
			db.Logger().Error("computing visitors", "domain", dom, "err", err)
		}
		sites[i]["visitors"] = vs
	}

	ctx := make(map[string]any)

	if len(sites) > 0 {
		ctx["sites"] = sites
	}
	db.HTML(w, sitesIndex, ctx)
}

func AddSnippet(db *db.Config, w http.ResponseWriter, r *http.Request) {
	tracker := fmt.Sprintf("%s/js/script.js", db.GetConfig().GetUrl())
	snippet := fmt.Sprintf(`<script defer data-domain=%q src=%q></script>`, db.CurrentSite().Domain, tracker)

	db.HTML(w, addSnippet, map[string]any{
		"snippet": template.HTML(snippet),
	})
}

func Unimplemented(db *db.Config, w http.ResponseWriter, r *http.Request) {
}

func RequireSiteAccess(h plug.Handler) plug.Handler {
	return func(db *db.Config, w http.ResponseWriter, r *http.Request) {
		domain := r.PathValue("domain")
		site := db.Ops().Site(domain)
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
			auth := r.URL.Query().Get("auth")
			if auth != "" {
				i, ok := slices.BinarySearchFunc(site.Shares, &v1.Share{Id: auth}, func(a, b *v1.Share) int {
					return cmp.Compare(a.Id, b.Id)
				})
				if !ok {
					db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
					return
				}
				share := site.Shares[i]

				if share.Password != nil {
					// verify shared link
					name := "shared-link-" + auth
					expires := db.LoadSharedLinkSession(r, name)
					if expires.Before(time.Now().UTC()) {
						dest := fmt.Sprintf("/v1/share/%s/authenticate/%s",
							url.PathEscape(site.Domain), auth)
						http.Redirect(w, r, dest, http.StatusFound)
						return
					}
					db.SetSite(site)
					h(db, w, r)
					return
				}
			}

			db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
			return
		}

		db.SetSite(site)
		h(db, w, r)
	}
}
