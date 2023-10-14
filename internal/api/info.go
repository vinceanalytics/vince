package api

import (
	"context"

	v1 "github.com/vinceanalytics/proto/gen/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/version"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (a *API) Version(ctx context.Context, _ *emptypb.Empty) (*v1.Build, error) {
	return &v1.Build{
		Version:  version.Build().String(),
		ServerId: config.Get(ctx).ServerId,
	}, nil
}
