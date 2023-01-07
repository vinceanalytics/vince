package vince

import "context"

type Vince struct {
	ts      *Tables
	session *SessionCache
}

type vinceKey struct{}

func (v *Vince) Witch(ctx context.Context) context.Context {
	return context.WithValue(ctx, vinceKey{}, v)
}

func From(ctx context.Context) *Vince {
	return ctx.Value(vinceKey{}).(*Vince)
}
