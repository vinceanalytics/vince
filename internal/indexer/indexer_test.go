package indexer

import (
	"testing"
	"time"

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/closter/events"
)

func TestIndexer(t *testing.T) {
	now, _ := time.Parse(time.RFC822, time.RFC822)
	now = now.UTC()
	ls := events.Samples(events.WithNow(func() time.Time { return now }))
	r := events.New(memory.DefaultAllocator).Write(ls)
	defer r.Release()
	idx := New()
	full, err := idx.Index(r)
	require.NoError(t, err)

	t.Run("Sets min,max timestamp", func(t *testing.T) {
		lo := time.UnixMilli(int64(full.Min())).UTC()
		require.Truef(t, now.Equal(lo), "now=%v lo=%v", now, lo)

		up := time.UnixMilli(ls.Items[len(ls.Items)-1].Timestamp)
		hi := time.UnixMilli(int64(full.Max())).UTC()
		require.Truef(t, up.Equal(hi), "now=%v hi=%v", up, hi)
	})
}
