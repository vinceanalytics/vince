package router

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vinceanalytics/vince/assets"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/a2"
	"github.com/vinceanalytics/vince/internal/api"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/metrics"
	"github.com/vinceanalytics/vince/internal/px"
	"github.com/vinceanalytics/vince/internal/tracker"
	"golang.org/x/crypto/bcrypt"
)

type Router struct {
	metrics http.Handler
	pprof   http.Handler
	a2      *a2.Server
}

func New(ctx context.Context) *Router {
	h := &Router{}
	h.metrics = promhttp.HandlerFor(metrics.Get(ctx), promhttp.HandlerOpts{})
	if config.Get(ctx).EnableProfile {
		h.pprof = http.HandlerFunc(pprof.Index)
	} else {
		h.pprof = http.HandlerFunc(NotFound)
	}
	h.a2 = a2.New(slog.Default())
	return h
}

func (h *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("x-frame-options", "SAMEORIGIN")
	w.Header().Set("x-xss-protection", "1; mode=block")
	w.Header().Set("x-content-type-options", "nosniff")
	w.Header().Set("x-download-options", "noopen")
	w.Header().Set("x-permitted-cross-domain-policies", "none")
	w.Header().Set("cross-origin-window-policy", "deny")

	if strings.HasPrefix(r.URL.Path, "/js/vince") {
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("cross-origin-resource-policy", "cross-origin")
		w.Header().Set("access-control-allow-origin", "*")
		w.Header().Set("cache-control", "public, max-age=86400, must-revalidate")
		tracker.Serve(w, r)
		return
	}
	if assets.Match(r.URL.Path) {
		assets.FS.ServeHTTP(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/debug/pprof/") {
		h.pprof.ServeHTTP(w, r)
		return
	}
	switch r.URL.Path {
	case "/metrics":
		h.metrics.ServeHTTP(w, r)
		return
	case "/api/event":
		api.Events(w, r)
		return
	case "/authorize":
		// resp := h.a2.NewResponse()
		// defer resp.Close()
		// if ar := h.a2.HandleAuthorizeRequest(resp, r); ar != nil {
		// 	ar.Authorized = true
		// 	h.a2.FinishAuthorizeRequest(resp, r, ar)
		// }
		// a2.OutputJSON(resp, w, r)
		return
	case "/token":
		resp := h.a2.NewResponse()
		defer resp.Close()
		ctx := r.Context()
		if ar := h.a2.HandleAccessRequest(resp, r); ar != nil {
			switch ar.Type {
			case a2.REFRESH_TOKEN:
				ar.Authorized = true
			case a2.PASSWORD:
				ar.Authorized = basic(ctx, ar)
			case a2.CLIENT_CREDENTIALS:
				ar.Authorized = true
			}
			h.a2.FinishAccessRequest(resp, r, ar)
		}
		a2.OutputJSON(resp, w, r)
		return
	case "/info":
		resp := h.a2.NewResponse()
		defer resp.Close()
		if ar := h.a2.HandleInfoRequest(resp, r); ar != nil {
			h.a2.FinishInfoRequest(resp, r, ar)
		}
		a2.OutputJSON(resp, w, r)
		return
	}
	NotFound(w, r)
}

func basic(ctx context.Context, ar *a2.AccessRequest) bool {
	var a v1.Account
	err := db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Account(ar.Username)
		return txn.Get(key, px.Decode(&a))
	})
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword(a.HashedPassword, []byte(ar.Password))
	return err == nil
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
