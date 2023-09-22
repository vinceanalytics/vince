package timeseries

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/must"
)

const DefaultSession = 15 * time.Minute

var sessionCache = must.Must(ristretto.NewCache(&ristretto.Config{
	NumCounters: 1e7,
	MaxCost:     2 << 20,
	BufferItems: 64,
}))("failed creating session cache")

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
	if o, ok := sessionCache.Get(e.ID); ok {
		s := o.(*entry.Entry)
		s.Update(e)
	} else {
		sessionCache.SetWithTTL(e.ID, e, 1, DefaultSession)
	}
	Block(ctx).WriteEntry(e)
}
