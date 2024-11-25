package web

import (
	"cmp"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/web/db"
	"golang.org/x/crypto/bcrypt"
)

func Share(db *db.Config, w http.ResponseWriter, r *http.Request) {
	site := db.CurrentSite()
	if site.Locked {
		db.HTML(w, statsLocked, nil)
		return
	}
	hasStats := db.Ops().SeenFirstStats(site.Domain)
	w.Header().Set("x-robots-tag", "noindex, nofollow")
	w.Header().Del("x-frame-options")
	auth := r.URL.Query().Get("auth")
	db.HTML(w, stats, map[string]any{
		"seen_first_stats": hasStats,
		"title":            "vince · " + site.Domain,
		"auth":             auth,
		"demo":             r.URL.Query().Get("demo") == "true",
	})
}

func ShareAuthForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, sharePassword, map[string]any{
		"domain": r.PathValue("domain"),
		"slug":   r.PathValue("slug"),
	})
}

func ShareAuth(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	domain := r.PathValue("domain")
	password := r.FormValue("password")
	site := db.Ops().Site(domain)
	if site == nil {
		db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
		return
	}
	slug := r.PathValue("slug")
	i, ok := slices.BinarySearchFunc(site.Shares, &v1.Share{Id: slug}, func(a, b *v1.Share) int {
		return cmp.Compare(a.Id, b.Id)
	})
	if !ok {
		db.HTMLCode(http.StatusNotFound, w, e404, map[string]any{})
		return
	}
	share := site.Shares[i]
	err := bcrypt.CompareHashAndPassword(share.Password, []byte(password))
	if err != nil {
		db.SaveCsrf(w)
		db.HTML(w, sharePassword, map[string]any{
			"domain": domain,
			"slug":   slug,
			"error":  "invalid pasword",
		})
		return
	}
	db.SaveSharedLinkSession(w, "shared-link-"+slug)
	dest := fmt.Sprintf("/v1/share/%s?auth=%s", url.PathEscape(domain), slug)
	http.Redirect(w, r, dest, http.StatusFound)

}
