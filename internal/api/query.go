package api

import (
	"context"
	"database/sql"

	apiv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/cmd/output"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/px"
	"github.com/vinceanalytics/vince/internal/query"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Query executes read only query. Assumes req has been validated.
func (a *API) Query(ctx context.Context, req *apiv1.QueryRequest) (*apiv1.QueryResponse, error) {
	params := make([]any, len(req.Params))
	for i := range params {
		params[i] = sql.Named(req.Params[i].Name,
			px.Interface(req.Params[i].Value))
	}
	db := query.GetInternalClient(ctx)
	start := core.Now(ctx)
	rows, err := db.Query(req.Query, params...)
	if err != nil {
		return nil, err
	}
	elapsed := core.Now(ctx).Sub(start)
	defer rows.Close()
	result, err := output.Build(rows)
	if err != nil {
		return nil, err
	}
	result.Elapsed = durationpb.New(elapsed)
	return result, nil
}
