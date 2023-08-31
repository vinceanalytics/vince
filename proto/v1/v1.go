package v1

import (
	"fmt"
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func (v *Query_Value) Interface() (val any) {
	switch e := v.Value.(type) {
	case *Query_Value_Number:
		val = e.Number
	case *Query_Value_Double:
		val = e.Double
	case *Query_Value_String_:
		val = e.String_
	case *Query_Value_Bool:
		val = e.Bool
	case *Query_Value_Timestamp:
		val = e.Timestamp.AsTime()
	}
	return
}

func NewQueryValue(v any) *Query_Value {
	switch e := v.(type) {
	case int64:
		return &Query_Value{
			Value: &Query_Value_Number{
				Number: e,
			},
		}
	case float64:
		return &Query_Value{
			Value: &Query_Value_Double{
				Double: e,
			},
		}
	case string:
		return &Query_Value{
			Value: &Query_Value_String_{
				String_: e,
			},
		}
	case bool:
		return &Query_Value{
			Value: &Query_Value_Bool{
				Bool: e,
			},
		}
	case time.Time:
		return &Query_Value{
			Value: &Query_Value_Timestamp{
				Timestamp: timestamppb.New(e),
			},
		}
	default:
		panic(fmt.Sprintf("unknown value type %#T", v))
	}
}

func (c Column) Index() int {
	if c <= Column_timestamp {
		return int(c)
	}
	return int(c - Column_browser)
}
