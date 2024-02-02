package api

import (
	"io"
	"net"
	"net/http"

	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/guard"
	"github.com/vinceanalytics/staples/staples/session"
	"google.golang.org/protobuf/encoding/protojson"
)

func ReceiveEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	ctx := r.Context()
	xg := guard.Get(ctx)
	if !xg.Allow() {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	var ev v1.Event
	b, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	err := protojson.Unmarshal(b, &ev)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !xg.Accept(ev.D) {
		w.Header().Set("x-vince-dropped", "1")
		w.WriteHeader(http.StatusOK)
		return
	}
	ev.Ip = remoteIP(r)
	ev.Ua = r.UserAgent()
	session.Get(ctx).Queue(ctx, &ev)
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
