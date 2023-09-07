package output

import (
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/query/v1"
)

func fromDBType(typ string) v1.QueryColum_DataType {
	switch typ {
	case "TEXT", "VARBINARY", "VARCHAR":
		return v1.QueryColum_STRING
	case "DOUBLE", "FLOAT":
		return v1.QueryColum_DOUBLE
	case "TIMESTAMP":
		return v1.QueryColum_TIMESTAMP
	case "UNSIGNED SMALLINT", "SMALLINT", "UNSIGNED BIGINT", "BIGINT":
		return v1.QueryColum_NUMBER
	default:
		return v1.QueryColum_UNKNOWN
	}
}
