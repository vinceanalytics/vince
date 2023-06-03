package api

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/health"
	"github.com/vinceanalytics/vince/render"
)

func Health(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, health.Get(r.Context()).Check(r.Context()))
}
