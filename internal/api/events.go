package api

import (
	"net"
	"net/http"

	"github.com/vinceanalytics/vince/internal/guard"
	"github.com/vinceanalytics/vince/internal/request"
	"github.com/vinceanalytics/vince/internal/session"
	"github.com/vinceanalytics/vince/internal/tenant"
)

func ReceiveEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", http.MethodPost)
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	ctx := r.Context()
	xg := guard.Get(ctx)
	if !xg.Allow() {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	ev := request.ReadEvent(w, r)
	if ev == nil {
		return
	}
	if !xg.Accept(ev.D) {
		w.Header().Set("x-vince-dropped", "1")
		w.WriteHeader(http.StatusOK)
		return
	}
	ev.Ip = remoteIP(r)
	ev.Ua = r.UserAgent()
	session.Get(ctx).Queue(ctx, tenant.Default, ev)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

var remoteIPHeaders = []string{
	"X-Real-IP", "X-Forwarded-For", "X-Client-IP",
}

func remoteIP(r *http.Request) string {
	var raw string
	for _, v := range remoteIPHeaders {
		if raw = r.Header.Get(v); raw != "" {
			break
		}
	}
	if raw == "" && r.RemoteAddr != "" {
		raw = r.RemoteAddr
	}
	var host string
	host, _, err := net.SplitHostPort(raw)
	if err != nil {
		host = raw
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "-"
	}
	return ip.String()
}
