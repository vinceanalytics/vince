package api

import (
	"net/http"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/v1"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/version"
)

func Version(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, &v1.Build{
		Version: version.Build().String(),
	})
}
