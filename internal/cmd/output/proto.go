package output

import (
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

func fromDBType(typ string) v1.Query_Colum_DataType {
	switch typ {
	case "TEXT", "VARBINARY", "VARCHAR":
		return v1.Query_Colum_STRING
	case "DOUBLE", "FLOAT":
		return v1.Query_Colum_DOUBLE
	case "TIMESTAMP":
		return v1.Query_Colum_TIMESTAMP
	case "UNSIGNED SMALLINT", "SMALLINT", "UNSIGNED BIGINT", "BIGINT":
		return v1.Query_Colum_NUMBER
	default:
		return v1.Query_Colum_UNKNOWN
	}
}
