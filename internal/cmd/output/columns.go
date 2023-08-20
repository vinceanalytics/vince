package output

import (
	"database/sql"
	"slices"

	v1 "github.com/vinceanalytics/vince/proto/v1"
)

func Build(rows *sql.Rows) (*v1.Query_Result, error) {
	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	r := &v1.Query_Result{}
	r.Columns = make([]*v1.Query_Colum, len(columns))
	for i := range r.Columns {
		r.Columns[i] = &v1.Query_Colum{
			Name:     columns[i].Name(),
			DataType: fromDBType(columns[i].DatabaseTypeName()),
		}
	}

	scans := make([]*v1.Query_Value, len(columns))
	clone := make([]any, len(columns))
	setValue := func(i int, v any) {
		scans[i] = v1.NewQueryValue(v)
	}

	for i := range clone {
		clone[i] = &Any{
			i:  i,
			cb: setValue,
		}
	}

	for rows.Next() {
		err = rows.Scan(clone...)
		if err != nil {
			return nil, err
		}
		r.Rows = append(r.Rows, &v1.Query_Row{
			Values: slices.Clone(scans),
		})
	}
	return r, rows.Err()
}

type Any struct {
	i  int
	cb func(int, any)
}

var _ sql.Scanner = (*Any)(nil)

func (a *Any) Scan(v any) error {
	a.cb(a.i, v)
	return nil
}
