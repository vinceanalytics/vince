package web

import (
	"errors"
	"net/http"

	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/web/db"
)

func Event(dba *db.Config, w http.ResponseWriter, r *http.Request) {
	// track how many requests hit this endpoint. Being public facing getting
	// insight on how much load we can handle is helpful metric.
	//
	// see Requests charts on system/stats endpoint
	ro2.Hit()

	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", http.MethodPost)
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")

	if err := dba.ProcessEvent(r); err != nil {
		if errors.Is(err, db.ErrDrop) {
			w.Header().Set("x-plausible-dropped", "1")
			w.WriteHeader(http.StatusAccepted)
			w.Write(ok)
			return
		}
		Json(w, map[string]any{"error": err.Error()}, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write(ok)
}

var ok = []byte("ok")
