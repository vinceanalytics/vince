package api

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/pj"
	"github.com/vinceanalytics/vince/internal/render"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

var client = &http.Client{}

func Bootstrap(w http.ResponseWriter, r *http.Request) {
	var b v1.Cluster_Bootstrap_Request
	err := pj.UnmarshalDefault(&b, r.Body)
	if err != nil {
		render.ERROR(w, http.StatusBadRequest)
		return
	}
	if b.Server == nil {
		render.ERROR(w, http.StatusBadRequest, "server is required")
		return
	}
	ctx := r.Context()
	o := config.Get(ctx)
	if o.ServerId != b.Server.Id {
		render.ERROR(w, http.StatusBadRequest, "invalid server id")
		return
	}
	if b.Server.Address == "" {
		render.ERROR(w, http.StatusBadRequest, "server address is required")
		return
	}
	res, err := client.Get(b.Server.Address)
	if err != nil {
		render.ERROR(w, http.StatusBadRequest, "invalid server address")
		return
	}
	res.Body.Close()
	render.JSON(w, http.StatusOK, &v1.Cluster_Bootstrap_Response{})
}
