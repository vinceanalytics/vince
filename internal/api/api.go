package api

import v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"

type API struct {
	v1.UnsafeVinceServer
}

var _ v1.VinceServer = (*API)(nil)
