package api

import (
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	clusterv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/cluster/v1"
	eventsv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/events/v1"
	goalsv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/goals/v1"
	importv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/import/v1"
	queryv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/query/v1"
	sitesv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/sites/v1"
	snippetsv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/snippets/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type API struct {
	v1.UnsafeVinceServer
	sitesv1.UnsafeSitesServer
	queryv1.UnsafeQueryServer
	goalsv1.UnsafeGoalsServer
	snippetsv1.UnsafeSnippetsServer
	clusterv1.UnsafeClusterServer
	eventsv1.UnsafeEventsServer
	importv1.UnsafeImportServer
}

var _ v1.VinceServer = (*API)(nil)

var e500 = status.Error(codes.Internal, "something went wrong")

func E500() error {
	return e500
}
