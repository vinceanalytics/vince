package api

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/version"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

func Version(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, &v1.Build{
		Version: version.Build().String(),
	})
}
