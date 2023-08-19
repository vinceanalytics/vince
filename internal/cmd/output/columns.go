package output

import (
	"database/sql"
	"fmt"
	"strings"
)

type Column struct {
	Name   string
	Values []any
}

func (c *Column) Format(w *strings.Builder, n int) {
	fmt.Fprint(w, c.Values[n])
}

var _ sql.Scanner = (*Column)(nil)

func (c *Column) Scan(src any) error {
	c.Values = append(c.Values, src)
	return nil
}

func Build(rows *sql.Rows) ([]*Column, error) {
	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	cols := make([]*Column, len(columns))
	scans := make([]any, len(columns))
	for i := range cols {
		cols[i] = &Column{
			Name: columns[i].Name(),
		}
		scans[i] = cols[i]
	}

	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			return nil, err
		}
	}
	return cols, rows.Err()
}
