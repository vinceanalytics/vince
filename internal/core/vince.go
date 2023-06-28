package core

import (
	"context"
	"net"
	"net/http"
	"time"
)

type httpListenerKey struct{}
type httpsListenerKey struct{}

type httpServerKey struct{}
type httpsServerKey struct{}

func SetHTTPListener(ctx context.Context, ls net.Listener) context.Context {
	return context.WithValue(ctx, httpListenerKey{}, ls)
}

func SetHTTPSListener(ctx context.Context, ls net.Listener) context.Context {
	return context.WithValue(ctx, httpsListenerKey{}, ls)
}

func GetHTTPListener(ctx context.Context) net.Listener {
	if v := ctx.Value(httpListenerKey{}); v != nil {
		return v.(net.Listener)
	}
	return nil
}

func GetHTTPSListener(ctx context.Context) net.Listener {
	if v := ctx.Value(httpsListenerKey{}); v != nil {
		return v.(net.Listener)
	}
	return nil
}

func SetHTTPServer(ctx context.Context, ls *http.Server) context.Context {
	return context.WithValue(ctx, httpServerKey{}, ls)
}

func SetHTTPSServer(ctx context.Context, ls *http.Server) context.Context {
	return context.WithValue(ctx, httpsServerKey{}, ls)
}

func GetHTTPServer(ctx context.Context) *http.Server {
	if v := ctx.Value(httpServerKey{}); v != nil {
		return v.(*http.Server)
	}
	return nil
}

func GetHTTPSServer(ctx context.Context) *http.Server {
	if v := ctx.Value(httpsServerKey{}); v != nil {
		return v.(*http.Server)
	}
	return nil
}

// NowFunc a functions that returns the current time.
type NowFunc func() time.Time

type nowKey struct{}

func SetNow(ctx context.Context, now NowFunc) context.Context {
	return context.WithValue(ctx, nowKey{}, now)
}

func GetNow(ctx context.Context) NowFunc {
	if v := ctx.Value(nowKey{}); v != nil {
		return v.(NowFunc)
	}
	return fallback
}

func fallback() time.Time {
	return time.Now().UTC()
}

func Now(ctx context.Context) time.Time {
	return GetNow(ctx)()
}

func Elapsed(ctx context.Context, since time.Time) time.Duration {
	return Now(ctx).Sub(since)
}
