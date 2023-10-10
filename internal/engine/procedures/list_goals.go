package procedures

import (
	"github.com/dolthub/go-mysql-server/sql"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/goals/v1"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/engine/session"
	"github.com/vinceanalytics/vince/internal/scopes"
)

func listGoal(ctx *sql.Context, domain string) (sql.RowIter, error) {
	req := v1.ListGoalsRequest{
		Domain: domain,
	}
	err := valid.Validate(&req)
	if err != nil {
		return nil, err
	}
	base := session.Get(ctx)
	if err = base.Allow(scopes.ListGoals); err != nil {
		return nil, err
	}
	res, err := (&api.API{}).ListGoals(base.Context(), &req)
	if err != nil {
		return nil, err
	}
	return goalsToRowIter(res.Goals...), nil
}
