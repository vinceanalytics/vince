package v1

import (
	"fmt"
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func (s *StoreKey) Parts() []string {
	if s.Namespace == "" {
		s.Namespace = "vince"
	}
	return []string{
		s.Namespace, s.Prefix.String(),
	}
}

func (s *Site_Key) Parts() []string {
	return append(s.Store.Parts(), s.Domain)
}

func (s *Block_Key) Parts() []string {
	return append(s.Store.Parts(), s.Kind.String(), s.Domain, s.Uid)
}

func (s *Account_Key) Parts() []string {
	return append(s.Store.Parts(), s.Name)
}

func (s *Token_Key) Parts() []string {
	return append(s.Store.Parts(), fmt.Sprint(s.Hash))
}

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
