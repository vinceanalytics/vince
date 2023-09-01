package api

import (
	"net/http"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/render"
)

func Health(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, &v1.Status{})
}
