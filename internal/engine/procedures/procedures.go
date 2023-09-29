package procedures

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

var Procedures = []sql.ExternalStoredProcedureDetails{}

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

// int64Schema returns a non-nullable schema with all columns as BIGINT.
func int64Schema(columnNames ...string) sql.Schema {
	sch := make(sql.Schema, len(columnNames))
	for i, colName := range columnNames {
		sch[i] = &sql.Column{
			Name:     colName,
			Type:     types.Int64,
			Nullable: false,
		}
	}
	return sch
}

// rowToIter returns a sql.RowIter with a single row containing the values passed in.
func rowToIter(vals ...interface{}) sql.RowIter {
	row := make(sql.Row, len(vals))
	for i, val := range vals {
		row[i] = val
	}
	return sql.RowsToRowIter(row)
}
