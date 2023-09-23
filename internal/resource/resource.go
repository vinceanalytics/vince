package resource

import (
	"context"
	"errors"
	"io"
)

type List []io.Closer

func (r List) Close() error {
	e := make([]error, 0, len(r))
	for i := len(r) - 1; i > 0; i-- {
		e = append(e, r[i].Close())
	}
	return errors.Join(e...)
}

func (r List) CloseWithGrace(ctx context.Context) error {
	e := make([]error, 0, len(r))
	for i := len(r) - 1; i > 0; i-- {
		if shut, ok := r[i].(Shutdown); ok {
			e = append(e, shut.Shutdown(ctx))
		} else {
			e = append(e, r[i].Close())
		}
	}
	return errors.Join(e...)
}

type Func func() error

func (f Func) Close() error {
	return f()
}

type Shutdown interface {
	Shutdown(context.Context) error
}

type Shut func(context.Context) error

func (f Shut) Shutdown(ctx context.Context) error {
	return f(ctx)
}
