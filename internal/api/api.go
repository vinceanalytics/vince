package api

import (
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	goalsv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/goals/v1"
	queryv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/query/v1"
	sitesv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
)

type API struct {
	v1.UnsafeVinceServer
	sitesv1.UnsafeSitesServer
	queryv1.UnsafeQueryServer
	goalsv1.UnsafeGoalsServer
}

var _ v1.VinceServer = (*API)(nil)
