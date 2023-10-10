package procedures

import (
	"github.com/dolthub/go-mysql-server/sql"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/goals/v1"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/engine/session"
	"github.com/vinceanalytics/vince/internal/scopes"
)

func deleteGoal(ctx *sql.Context, domain, name string) (sql.RowIter, error) {
	req := v1.DeleteGoalRequest{
		Name:   name,
		Domain: domain,
	}
	err := valid.Validate(&req)
	if err != nil {
		return nil, err
	}
	base := session.Get(ctx)
	if err = base.Allow(scopes.DeleteGoal); err != nil {
		return nil, err
	}
	_, err = (&api.API{}).DeleteGoal(base.Context(), &req)
	if err != nil {
		return nil, err
	}
	return rowToIter("ok"), nil
}
