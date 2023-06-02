package api

import (
	"net/http"
	"runtime/debug"

	"github.com/vinceanalytics/vince/render"
)

func Version(w http.ResponseWriter, r *http.Request) {
	build, _ := debug.ReadBuildInfo()
	render.JSON(w, http.StatusOK, build)
}
