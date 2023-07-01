package timeseries

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vinceanalytics/vince/pkg/log"
)

func SaveSystem(ctx context.Context) {
	m, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		log.Get().Err(err).Msg("failed to gather stats")
		return
	}
	_ = m
}
