package api

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/pkg/version"
)

func Version(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, map[string]any{
		"version": version.Build().String(),
	})
}
