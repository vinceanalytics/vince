package output

import (
	"database/sql"
	"fmt"
	"time"

	sqld "github.com/dolthub/go-mysql-server/sql"
	v1 "github.com/vinceanalytics/proto/gen/go/vince/query/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Build(rows *sql.Rows) (*v1.QueryResponse, error) {
	columns, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	r := &v1.QueryResponse{}
	r.Columns = make([]*v1.QueryColum, len(columns))
	for i := range r.Columns {
		r.Columns[i] = &v1.QueryColum{
			Name:     columns[i].Name(),
			DataType: fromDBType(columns[i].DatabaseTypeName()),
		}
	}

	row := make([]*v1.QueryValue, len(columns))
	clone := make([]any, len(columns))

	for i := range clone {
		switch r.Columns[i].DataType {
		case v1.QueryColum_NUMBER:
			var v int64
			clone[i] = &v
		case v1.QueryColum_DOUBLE:
			var v float64
			clone[i] = &v
		case v1.QueryColum_BOOL:
			var v bool
			clone[i] = &v
		case v1.QueryColum_STRING:
			var v string
			clone[i] = &v
		case v1.QueryColum_TIMESTAMP:
			var v Time
			clone[i] = &v
		}
	}

	for rows.Next() {
		err = rows.Scan(clone...)
		if err != nil {
			return nil, err
		}
		row = row[:0]
		for i := range clone {
			switch e := clone[i].(type) {
			case *string:
				row = append(row, &v1.QueryValue{
					Value: &v1.QueryValue_String_{
						String_: *e,
					},
				})
			case *int64:
				row = append(row, &v1.QueryValue{
					Value: &v1.QueryValue_Number{
						Number: *e,
					},
				})
			case *float64:
				row = append(row, &v1.QueryValue{
					Value: &v1.QueryValue_Double{
						Double: *e,
					},
				})
			case *bool:
				row = append(row, &v1.QueryValue{
					Value: &v1.QueryValue_Bool{
						Bool: *e,
					},
				})
			case *Time:
				row = append(row, &v1.QueryValue{
					Value: &v1.QueryValue_Timestamp{
						Timestamp: timestamppb.New(
							e.Time,
						),
					},
				})
			}
		}
		r.Rows = append(r.Rows, &v1.QueryRow{
			Values: row,
		})
	}
	return r, rows.Err()
}

type Time struct {
	Time time.Time
}

var _ sql.Scanner = (*Time)(nil)

func (t *Time) Scan(src any) error {
	switch e := src.(type) {
	case time.Time:
		t.Time = e
	case []byte:
		var err error
		t.Time, err = time.Parse(sqld.TimestampDatetimeLayout, string(e))
		if err != nil {
			return err
		}
	case string:
		var err error
		t.Time, err = time.Parse(sqld.TimestampDatetimeLayout, e)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("can't convert %T to time.Time", e)
	}
	return nil
}
