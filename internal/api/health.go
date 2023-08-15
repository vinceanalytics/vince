package api

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/render"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

func Health(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, &v1.Status{})
}
