package worker

import (
	"context"

	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/events"
	"github.com/vinceanalytics/vince/internal/timeseries"
)

type requestsKey struct{}

type RequestBuffer struct {
	buf chan *entry.Request
}

func (r *RequestBuffer) Close() error {
	close(r.buf)
	return nil
}

func (r *RequestBuffer) Accept(req *entry.Request) {
	r.buf <- req
}

func SetupRequestsBuffer(ctx context.Context) (context.Context, *RequestBuffer) {
	r := &RequestBuffer{
		buf: make(chan *entry.Request, 4<<10),
	}
	return context.WithValue(ctx, requestsKey{}, r), r
}

func GetBuff(ctx context.Context) *RequestBuffer {
	return ctx.Value(requestsKey{}).(*RequestBuffer)
}

func ProcessRequests(ctx context.Context) {
	b := GetBuff(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-b.buf:
			e, err := events.Parse(r, core.Now(ctx))
			if err != nil {
			} else {
				timeseries.Register(ctx, e)
			}
			r.Release()
		}
	}
}

func Submit(ctx context.Context, r *entry.Request) {
	GetBuff(ctx).Accept(r)
}
