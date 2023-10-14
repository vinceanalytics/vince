package procedures

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	v1 "github.com/vinceanalytics/proto/gen/go/vince/goals/v1"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/engine/session"
	"github.com/vinceanalytics/vince/internal/scopes"
)

func createGoal(ctx *sql.Context, domain, name string, typ string, value string) (sql.RowIter, error) {
	var gt v1.Goal_Type
	switch strings.ToLower(typ) {
	case "event":
		gt = v1.Goal_EVENT
	case "path":
		gt = v1.Goal_PATH
	}
	req := v1.CreateGoalRequest{
		Name:   name,
		Domain: domain,
		Type:   gt,
		Value:  value,
	}
	err := valid.Validate(&req)
	if err != nil {
		return nil, err
	}
	base := session.Get(ctx)
	if err = base.Allow(scopes.CreateGoal); err != nil {
		return nil, err
	}
	_, err = (&api.API{}).CreateGoal(base.Context(), &req)
	if err != nil {
		return nil, err
	}
	return rowToIter("ok"), nil
}

var goalSchema = sql.Schema{
	{Name: "name", Type: types.Text},
	{Name: "type", Type: types.Text},
	{Name: "value", Type: types.Text},
	{Name: "created_at", Type: types.Timestamp},
}

func goalsToRowIter(goals ...*v1.Goal) sql.RowIter {
	rows := make([]sql.Row, len(goals))
	for i := range goals {
		rows[i] = goalToRow(goals[i])
	}
	return sql.RowsToRowIter(rows...)
}

func goalToRow(g *v1.Goal) sql.Row {
	o := make(sql.Row, 4)
	o[0] = g.Name
	o[1] = g.Type.String()
	o[2] = g.Value
	o[3] = g.CreatedAt.AsTime()
	return o
}
