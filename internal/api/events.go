package api

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/events"
	"github.com/vinceanalytics/vince/internal/gate"
	"github.com/vinceanalytics/vince/internal/remoteip"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/entry"
	"github.com/vinceanalytics/vince/pkg/log"
)

// Events accepts events payloads from vince client script.
func Events(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("X-Content-Type-Options", "nosniff")
	xlg := log.Get()
	req := entry.NewRequest()
	defer req.Release()

	err := req.Parse(r.Body)
	if err != nil {
		xlg.Err(err).Msg("Failed decoding json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.IP = remoteip.Get(r)
	req.UserAgent = r.UserAgent()

	e, err := events.Parse(req, core.Now(r.Context()))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pass := gate.Check(r.Context(), req.Domain)
	if !pass {
		w.Header().Set("x-vince-dropped", "1")
		w.WriteHeader(http.StatusAccepted)
		return
	}
	timeseries.Register(r.Context(), e)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
