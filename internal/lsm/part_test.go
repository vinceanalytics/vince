package lsm

import (
	"testing"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/cluster/events"
	"github.com/vinceanalytics/vince/internal/indexer"
)

func TestPartStore(t *testing.T) {
	ps := NewPartStore(memory.NewGoAllocator())
	defer ps.Release()

	now := events.Now()()
	// Three records starting 1 hour apart
	first := events.SampleRecord(events.WithNow(nowFunc(now)), events.WithStep(time.Minute))
	second := events.SampleRecord(events.WithNow(nowFunc(now.Add(time.Hour))), events.WithStep(time.Minute))
	third := events.SampleRecord(events.WithNow(nowFunc(now.Add(2*time.Hour))), events.WithStep(time.Minute))

	p1 := mustPart(t, first)
	p2 := mustPart(t, second)
	p3 := mustPart(t, third)

	ps.Add(p1)
	ps.Add(p2)
	ps.Add(p3)

	t.Run("Size", func(t *testing.T) {
		wantSize := p1.Size() + p2.Size() + p3.Size()
		require.Equal(t, wantSize, ps.Size())
	})
	t.Run("Compact", func(t *testing.T) {
		old := ps.Size()
		stats := ps.Compact(func(r arrow.Record) {
			defer r.Release()
			wantRows := first.NumRows() + second.NumRows() + third.NumRows()
			require.Equal(t, wantRows, r.NumRows())
		})
		require.Equal(t, old, stats.OldSize)
		wantNodes := 3
		require.Equal(t, wantNodes, stats.CompactedNodesCount)

		require.Zero(t, ps.Size())
	})
}

func mustPart(t *testing.T, r arrow.Record) *RecordPart {
	t.Helper()
	idx, err := indexer.New().Index(r)
	if err != nil {
		t.Fatal(err)
	}
	return NewPart(r, idx)
}
func nowFunc(now time.Time) func() time.Time {
	return func() time.Time { return now }
}
