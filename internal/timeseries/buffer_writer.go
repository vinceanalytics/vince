package timeseries

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/pkg/entry"
)

const DefaultSession = 15 * time.Minute

// Register records entry e. Tracks duration between page navigation by using a
// cache. Entries are cached by the user hash. By default the entry is kept for
// 15 mins before it is discarded.
//
// By caching , it allows accurate tracking of page navigation. Giving us
// - entry page
// - exit page
// - duration on page
// - bounce rate
func Register(ctx context.Context, e *entry.Entry) {
	e.Hit()
	x := caches.Session(ctx)
	cacheKey := e.ID
	if o, ok := x.Get(cacheKey); ok {
		s := o.(*entry.Entry)
		s.Update(e)
	} else {
		x.SetWithTTL(e.ID, e.Clone(), 1, DefaultSession)
	}
	Block(ctx).WriteEntry(e)
}
