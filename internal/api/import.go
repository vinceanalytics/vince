package api

import (
	"context"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/import/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ v1.ImportServer = (*API)(nil)

func (API) Import(context.Context, *v1.ImportRequest) (*v1.ImportResponse, error) {
	return nil, status.Error(codes.Unimplemented, "imports are not supported yet")
}
