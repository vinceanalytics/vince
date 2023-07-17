package timeseries

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/pkg/entry"
)

const DefaultSession = 15 * time.Minute

func Register(ctx context.Context, e *entry.Entry) {
	e.Hit()
	x := caches.Session(ctx)
	cacheKey := e.ID
	if o, ok := x.Get(cacheKey); ok {
		s := o.(*entry.Entry)
		s.Update(e)
		defer e.Release()
	} else {
		x.SetWithTTL(e.ID, e, 1, DefaultSession)
	}
	Block(ctx).WriteEntry(e)
}
