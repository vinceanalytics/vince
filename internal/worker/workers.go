package worker

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/health"
	"github.com/vinceanalytics/vince/internal/log"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"google.golang.org/protobuf/types/known/durationpb"
)

func Periodic(
	ctx context.Context,
	ping *health.Ping,
	interval time.Duration,
	warm bool, // if true calls work before starting the loop
	work func(context.Context, time.Duration)) func() error {
	tick := time.NewTicker(interval)
	log.Get().Debug().Str("worker", ping.Key).
		Str("every", interval.String()).
		Msg("started")
	if warm {
		work(ctx, interval)
	}
	return func() error {
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case pong := <-ping.Channel:
				pong()
			case <-tick.C:
				work(ctx, interval)
			}
		}
	}
}

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
