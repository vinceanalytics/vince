package core

import (
	"context"
	"net"
	"net/http"
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
