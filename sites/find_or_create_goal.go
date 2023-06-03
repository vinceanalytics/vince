package sites

import (
	"encoding/json"
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/render"
)

type GoalRequest struct {
	SiteID    string `json:"site_id"`
	GoalType  string `json:"goal_type"`
	PagePath  string `json:"page_path,omitempty"`
	EventName string `json:"event_name,omitempty"`
}

func (g *GoalRequest) valid() string {
	if g.SiteID == "" {
		return "site_id is required"
	}
	if g.GoalType == "" {
		return "goal_type is required"
	}
	switch g.GoalType {
	case "event":
		if g.EventName == "" {
			return "event_name is required"
		}
	case "page":
		if g.PagePath == "" {
			return "page_path is required"
		}
	}
	return ""
}

func FindOrCreateGoals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var g GoalRequest
	json.NewDecoder(r.Body).Decode(&g)
	if v := g.valid(); v != "" {
		render.JSON(w, http.StatusBadRequest, map[string]any{
			"error": v,
		})
		return
	}
	site := models.SiteFor(ctx,
		models.GetUser(ctx).ID,
		g.SiteID,
		"owner", "admin",
	)
	if site == nil {
		render.JSON(w, http.StatusNotFound, map[string]any{
			"error": http.StatusText(http.StatusNotFound),
		})
		return
	}
	goal := models.CreateGoal(ctx, site.Domain, g.EventName, g.PagePath)
	if goal == nil {
		render.JSON(w, http.StatusBadRequest, map[string]any{
			"error": http.StatusText(http.StatusBadRequest),
		})
		return
	}
	render.JSON(w, http.StatusOK, goal)
}
