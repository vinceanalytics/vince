package render

import (
	"fmt"
	"net/http"

	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
)

func JSON(w http.ResponseWriter, code int, m proto.Message) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(must.Must(pj.Marshal(m))(
		"failed encoding json object",
	))
}

func ERROR(w http.ResponseWriter, code int, msg ...any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	err := &v1.Error{
		Code:  int32(code),
		Error: http.StatusText(code),
	}
	if len(msg) != 0 {
		err.Error = fmt.Sprint(msg...)
	}
	data := must.Must(pj.Marshal(err))(
		"failed encoding error message",
	)
	println(len(data))
	fmt.Println(w.Write(data))
}
