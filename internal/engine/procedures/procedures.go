package procedures

import (
	"github.com/bufbuild/protovalidate-go"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/vinceanalytics/vince/internal/must"
)

var Procedures = []sql.ExternalStoredProcedureDetails{
	{Name: "add_site", Schema: stringSchema("status"), Function: addSite},
	{Name: "add_site", Schema: stringSchema("status"), Function: addSiteWithDescription},
	{Name: "get_site", Schema: siteSchema(), Function: getSite},
	{Name: "list_sites", Schema: siteSchema(), Function: listSites},
	{Name: "delete_site", Schema: stringSchema("status"), Function: deleteSite},
	// goals api
	{Name: "create_goal", Schema: stringSchema("status"), Function: createGoal},
	{Name: "get_goal", Schema: goalSchema, Function: getGoal},
	{Name: "list_goals", Schema: goalSchema, Function: listGoal},
	{Name: "delete_goal", Schema: stringSchema("status"), Function: deleteGoal},

	// build
	{Name: "get_info", Schema: infoSchema, Function: getInfo},
}

var valid = must.Must(protovalidate.New(protovalidate.WithFailFast(true)))("failed creating validator")

// stringSchema returns a non-nullable schema with all columns as LONGTEXT.
func stringSchema(columnNames ...string) sql.Schema {
	sch := make(sql.Schema, len(columnNames))
	for i, colName := range columnNames {
		sch[i] = &sql.Column{
			Name:     colName,
			Type:     types.LongText,
			Nullable: false,
		}
	}
	return sch
}

// rowToIter returns a sql.RowIter with a single row containing the values passed in.
func rowToIter(vals ...interface{}) sql.RowIter {
	return sql.RowsToRowIter(vals)
}
