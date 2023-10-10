package procedures

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/engine/session"
)

func getInfo(ctx *sql.Context) (sql.RowIter, error) {
	base := session.Get(ctx)
	res, err := (&api.API{}).Version(base.Context(), nil)
	if err != nil {
		return nil, err
	}
	o := make(sql.Row, 2)
	o[0] = res.ServerId
	o[1] = res.Version
	return sql.RowsToRowIter(o), nil
}

var infoSchema = sql.Schema{
	{Name: "server_id", Type: types.Text},
	{Name: "version", Type: types.Text},
}
