package api

import (
	"context"
	"net/http"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/version"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Version(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, &v1.Build{
		Version: version.Build().String(),
	})
}

func (a *API) Version(context.Context, *emptypb.Empty) (*v1.Build, error) {
	return &v1.Build{
		Version: version.Build().String(),
	}, nil
}
