package core

import (
	"context"
	"net"
	"time"

	v1 "github.com/vinceanalytics/proto/gen/go/vince/config/v1"
)

type httpListenerKey struct{}

func SetHTTPListener(ctx context.Context, ls net.Listener) context.Context {
	return context.WithValue(ctx, httpListenerKey{}, ls)
}

func GetHTTPListener(ctx context.Context) net.Listener {
	if v := ctx.Value(httpListenerKey{}); v != nil {
		return v.(net.Listener)
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

type authKey struct{}

func SetAuth(ctx context.Context, a *v1.Client_Auth) context.Context {
	return context.WithValue(ctx, authKey{}, a)
}

func GetAuth(ctx context.Context) *v1.Client_Auth {
	return ctx.Value(authKey{}).(*v1.Client_Auth)
}
