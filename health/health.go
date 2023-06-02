package health

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/vinceanalytics/vince/render"
)

type Component interface {
	Name() string
	Check(context.Context) bool
	Clone() Component
	io.Closer
}

var _ Component = (*Base)(nil)
var _ Component = (*Ping)(nil)

type Base struct {
	Key       string
	CheckFunc func(ctx context.Context) bool
}

func (b Base) Name() string {
	return b.Key
}

func (b Base) Check(ctx context.Context) bool {
	return b.CheckFunc(ctx)
}

func (b Base) Close() error {
	return nil
}

func (b Base) Clone() Component {
	return b
}

type Health []Component

func (h Health) Check(ctx context.Context) map[string]bool {
	o := make(map[string]bool)
	for _, x := range h {
		o[x.Name()] = x.Clone().Check(ctx)
	}
	return o
}

func (h Health) Close() error {
	e := make([]error, len(h))
	for i, x := range h {
		e[i] = x.Close()
	}
	return errors.Join(e...)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, http.StatusOK, Get(r.Context()).Check(r.Context()))
}

type healthKey struct{}

func Set(ctx context.Context, h Health) context.Context {
	return context.WithValue(ctx, healthKey{}, h)
}

func Get(ctx context.Context) Health {
	return ctx.Value(healthKey{}).(Health)
}

type PingChannel chan func()

type Ping struct {
	Key     string
	Channel PingChannel
}

func NewPing(name string) *Ping {
	return &Ping{
		Key:     name,
		Channel: make(PingChannel),
	}
}

func (p Ping) Name() string {
	return p.Key
}

func (p *Ping) Check(ctx context.Context) bool {
	o := make(chan struct{}, 1)
	defer close(o)
	p.Channel <- func() {
		o <- struct{}{}
	}
	ctx, _ = context.WithTimeout(ctx, 10*time.Millisecond)
	select {
	case <-ctx.Done():
		return false
	case <-o:
		return true
	}
}

func (p *Ping) Close() error {
	close(p.Channel)
	return nil
}

func (p *Ping) Clone() Component {
	return &Ping{
		Key:     p.Key,
		Channel: p.Channel,
	}
}
