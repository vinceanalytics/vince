package sites

import (
	"encoding/json"
	"net/http"

	"github.com/vinceanalytics/vince/config"
	"github.com/vinceanalytics/vince/models"
	"github.com/vinceanalytics/vince/render"
)

type SharedLinkRequest struct {
	SiteID   string `json:"site_id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (g *SharedLinkRequest) valid() string {
	if g.SiteID == "" {
		return "site_id is required"
	}
	if g.Name == "" {
		return "name is required"
	}
	return ""
}

func FindOrCreateSharedLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var g SharedLinkRequest
	json.NewDecoder(r.Body).Decode(&g)
	if v := g.valid(); v != "" {
		render.JSON(w, http.StatusBadRequest, map[string]any{
			"error": v,
		})
		return
	}
	site := models.SiteByDomain(ctx, g.SiteID)
	if site == nil {
		render.JSON(w, http.StatusNotFound, map[string]any{
			"error": http.StatusText(http.StatusNotFound),
		})
		return
	}
	shared := models.GetSharedLink(ctx, site.ID, g.Name)
	if shared == nil {
		shared = models.CreateSharedLink(ctx, shared.ID, g.Name, g.Password)
	}
	render.JSON(w, http.StatusOK, map[string]any{
		"name": shared.Name,
		"url":  models.SharedLinkURL(config.Get(ctx).Url, site, shared),
	})
}
