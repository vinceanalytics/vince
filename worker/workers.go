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

func StartEventWriter(
	ctx context.Context,
	source <-chan *timeseries.Event,
	ts *timeseries.Tables,
	wg *sync.WaitGroup,
	exit func(),
) health.Component {
	wg.Add(1)
	h := health.NewPing("event_writer_worker")
	go eventWriter(ctx, source, ts, wg, h.Channel, exit)
	return h
}

func eventWriter(
	ctx context.Context,
	source <-chan *timeseries.Event,
	ts *timeseries.Tables,
	wg *sync.WaitGroup,
	h health.PingChannel,
	exit func(),
) {
	ev := func() *zerolog.Event {
		return log.Get(ctx).Debug().Str("worker", "event_writer")
	}
	ev().Msg("start")
	defer func() {
		ev().Msg("exit")
		defer wg.Done()
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	events := make([]*timeseries.Event, 0, MAX_BUFFER_SIZE)
	flush := func() {
		count := len(events)
		if count == 0 {
			return
		}
		_, err := ts.WriteEvents(events)
		if err != nil {
			ev().Err(err).Msg("saving events")
			exit()
			return
		}
		events = events[:0]
	}
	for {
		select {
		case <-ctx.Done():
			return
		case f := <-h:
			f()
		case <-ticker.C:
			flush()
		case e := <-source:
			events = append(events, e)
			if len(events) == MAX_BUFFER_SIZE {
				flush()
			}
		}
	}
}

func StartSessionWriter(
	ctx context.Context,
	source <-chan *timeseries.Session,
	ts *timeseries.Tables,
	wg *sync.WaitGroup,
	exit func(),
) health.Component {
	wg.Add(1)
	h := health.NewPing("session_writer_worker")
	go sessionWriter(ctx, source, ts, wg, h.Channel, exit)
	return h
}

func sessionWriter(
	ctx context.Context,
	source <-chan *timeseries.Session,
	ts *timeseries.Tables,
	wg *sync.WaitGroup,
	h health.PingChannel,
	exit func(),
) {
	ev := func() *zerolog.Event {
		return log.Get(ctx).Debug().Str("worker", "session_writer")
	}
	ev().Msg("start")
	defer func() {
		ev().Msg("exit")
		defer wg.Done()
	}()
	sessions := make([]*timeseries.Session, 0, MAX_BUFFER_SIZE)
	flush := func() {
		count := len(sessions)
		if count == 0 {
			return
		}
		_, err := ts.WriteSessions(sessions)
		if err != nil {
			ev().Err(err).Msg("saving sessions")
			exit()
			return
		}
		sessions = sessions[:0]
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
		case sess := <-source:
			sessions = append(sessions, sess)
			if len(sessions) == MAX_BUFFER_SIZE {
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
				ev().Msg("failed archiving events")
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
