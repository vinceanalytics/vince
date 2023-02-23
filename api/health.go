package api

import (
	"net/http"

	"github.com/gernest/vince/health"
	"github.com/gernest/vince/render"
)

func Health(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, health.Get(r.Context()).Check(r.Context()))
}
