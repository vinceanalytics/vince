package worker

import (
	"context"
	"sync"
	"time"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/health"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/timeseries"
	"github.com/rs/zerolog"
)

const MAX_BUFFER_SIZE = 4098

func Flush[T any](
	ctx context.Context,
	name string,
	source <-chan T,
	write func([]T) (int, error),
	wg *sync.WaitGroup,
	exit func(),
) health.Component {
	wg.Add(1)
	h := health.NewPing(name)
	go flushInternal(ctx, name, source, write, wg, h.Channel, exit)
	return h
}

func flushInternal[T any](
	ctx context.Context,
	name string,
	source <-chan T,
	write func([]T) (int, error),
	wg *sync.WaitGroup,
	h health.PingChannel,
	exit func(),
) {
	defer func() {
		defer wg.Done()
	}()
	ls := make([]T, 0, MAX_BUFFER_SIZE)
	flush := func() {
		count := len(ls)
		if count == 0 {
			return
		}
		_, err := write(ls)
		if err != nil {
			exit()
			return
		}
		ls = ls[:0]
	}
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case f := <-h:
			f()
		case <-ticker.C:
			flush()
		case v := <-source:
			ls = append(ls, v)
			if len(ls) == MAX_BUFFER_SIZE {
				flush()
			}
		}
	}

}

func StartSeriesArchive(
	ctx context.Context,
	ts *timeseries.Tables,
	wg *sync.WaitGroup,
	exit func(),
) health.Component {
	wg.Add(1)
	h := health.NewPing("series_archive_writer_worker")
	go seriesArchive(ctx, ts, wg, h.Channel, exit)
	return h
}

func seriesArchive(
	ctx context.Context,
	ts *timeseries.Tables,
	wg *sync.WaitGroup,
	h health.PingChannel,
	exit func(),
) {
	ev := func() *zerolog.Event {
		return log.Get(ctx).Debug().Str("worker", "series_archive")
	}
	interval := config.Get(ctx).FlushInterval.AsDuration()
	ev().Dur("interval", interval).Msg("start")
	defer func() {
		ev().Msg("exit")
		defer wg.Done()
	}()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case f := <-h:
			f()
		case <-ticker.C:
			n, err := ts.ArchiveEvents()
			if err != nil {
				ev().Err(err).Msg("failed archiving events")
				exit()
				return
			}
			ev().Int64("size", n).Msg("archiving events")
			n, err = ts.ArchiveSessions()
			if err != nil {
				ev().Err(err).Msg("failed archiving sessions")
				exit()
				return
			}
			ev().Int64("size", n).Msg("archiving sessions")
		}
	}
}
