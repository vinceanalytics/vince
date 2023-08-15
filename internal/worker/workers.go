package worker

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/timeseries"
	"google.golang.org/protobuf/types/known/durationpb"
)

// SaveBuffers persists collected Entry Buffers to the timeseries storage.
func SaveBuffers(ctx context.Context, interval *durationpb.Duration) {
	ts := time.NewTicker(interval.AsDuration())
	defer ts.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ts.C:
			timeseries.Save(ctx)
		}
	}
}
