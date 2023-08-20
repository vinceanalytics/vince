package query

import (
	"testing"
	"time"

	"github.com/vinceanalytics/vince/internal/pj"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestYay(t *testing.T) {
	a, _ := pj.MarshalIndent(&v1.Query_Value{
		Value: &v1.Query_Value_Number{},
	})
	b, _ := pj.MarshalIndent(&v1.Query_Value{
		Value: &v1.Query_Value_Double{},
	})
	c, _ := pj.MarshalIndent(&v1.Query_Value{
		Value: &v1.Query_Value_String_{},
	})
	d, _ := pj.MarshalIndent(&v1.Query_Value{
		Value: &v1.Query_Value_Bool{},
	})
	e, _ := pj.MarshalIndent(&v1.Query_Value{
		Value: &v1.Query_Value_Timestamp{
			Timestamp: timestamppb.New(time.Time{}),
		},
	})
	println(string(a))
	println(string(b))
	println(string(c))
	println(string(d))
	println(string(e))
	t.Error()
}
