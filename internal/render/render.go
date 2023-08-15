package render

import (
	"encoding/json"
	"net/http"

	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	"google.golang.org/protobuf/proto"
)

func JSON(w http.ResponseWriter, code int, m proto.Message) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(must.Must(pj.Marshal(m))())
}

type err struct {
	Error any `json:"error"`
}

func ERROR(w http.ResponseWriter, code int, msg ...any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	if len(msg) == 0 {
		msg = []any{http.StatusText(code)}
	}
	json.NewEncoder(w).Encode(err{Error: msg[0]})
}
