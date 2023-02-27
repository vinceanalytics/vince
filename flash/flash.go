package flash

import "context"

type flashKey struct{}

func Set(ctx context.Context, f *Flash) context.Context {
	return context.WithValue(ctx, flashKey{}, f)
}

func Get(ctx context.Context) *Flash {
	if f := ctx.Value(flashKey{}); f != nil {
		return f.(*Flash)
	}
	return nil
}

type Flash struct {
	Success []string
	Failure []string
}
