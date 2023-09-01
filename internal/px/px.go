package px

import (
	"fmt"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func Interface(v *v1.Query_Value) (val any) {
	switch e := v.Value.(type) {
	case *v1.Query_Value_Number:
		val = e.Number
	case *v1.Query_Value_Double:
		val = e.Double
	case *v1.Query_Value_String_:
		val = e.String_
	case *v1.Query_Value_Bool:
		val = e.Bool
	case *v1.Query_Value_Timestamp:
		val = e.Timestamp.AsTime()
	}
	return
}

func NewQueryValue(v any) *v1.Query_Value {
	switch e := v.(type) {
	case int64:
		return &v1.Query_Value{
			Value: &v1.Query_Value_Number{
				Number: e,
			},
		}
	case float64:
		return &v1.Query_Value{
			Value: &v1.Query_Value_Double{
				Double: e,
			},
		}
	case string:
		return &v1.Query_Value{
			Value: &v1.Query_Value_String_{
				String_: e,
			},
		}
	case bool:
		return &v1.Query_Value{
			Value: &v1.Query_Value_Bool{
				Bool: e,
			},
		}
	case time.Time:
		return &v1.Query_Value{
			Value: &v1.Query_Value_Timestamp{
				Timestamp: timestamppb.New(e),
			},
		}
	default:
		panic(fmt.Sprintf("unknown value type %#T", v))
	}
}

func ColumnIndex(c storev1.Column) int {
	if c <= storev1.Column_timestamp {
		return int(c)
	}
	return int(c - storev1.Column_browser)
}
