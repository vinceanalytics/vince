package sites

import (
	"encoding/json"
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
)

type CreateSiteRequest struct {
	Domain string `json:"domain"`
	Public bool   `json:"public,omitempty"`
}

type Error struct {
	Error string `json:"error"`
}

func CreateSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	var data CreateSiteRequest
	json.NewDecoder(r.Body).Decode(&data)
	domain, bad := models.ValidateSiteDomain(ctx, data.Domain)
	if bad != "" {
		render.JSON(w, http.StatusBadRequest, Error{
			Error: "domain: " + bad,
		})
		return
	}
	if !models.CreateSite(ctx, u, domain, data.Public) {
		render.JSON(w, http.StatusBadRequest, Error{
			Error: http.StatusText(http.StatusInternalServerError),
		})
		return
	}
	site := models.SiteByDomain(ctx, domain)
	render.JSON(w, http.StatusOK, site)
}
