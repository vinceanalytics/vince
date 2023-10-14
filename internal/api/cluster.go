package api

import (
	"context"

	clusterv1 "github.com/vinceanalytics/proto/gen/go/vince/cluster/v1"
)

var _ clusterv1.ClusterServer = (*API)(nil)

func (a *API) ApplyCluster(context.Context, *clusterv1.ApplyClusterRequest) (*clusterv1.ApplyClusterResponse, error) {
	return nil, nil
}

func (a *API) GetCluster(context.Context, *clusterv1.GetClusterRequest) (*clusterv1.GetClusterResponse, error) {
	return nil, nil
}
