package system

import (
	"math"
	"sync/atomic"
	"time"
)

type Counter struct {
	Timestamp time.Time `parquet:"timestamp,dict,zstd"`
	Name      string    `parquet:"name,dict,zstd"`
	Value     float64   `parquet:"value,dict,zstd"`
}

type counterMetric struct {
	valBits uint64
	valInt  uint64
	name    string
}

func (c *counterMetric) Add(v float64) {
	ival := uint64(v)
	if float64(ival) == v {
		atomic.AddUint64(&c.valInt, ival)
		return
	}
	for {
		oldBits := atomic.LoadUint64(&c.valBits)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + v)
		if atomic.CompareAndSwapUint64(&c.valBits, oldBits, newBits) {
			return
		}
	}
}

func (c *counterMetric) Inc() {
	atomic.AddUint64(&c.valInt, 1)
}

func (c *counterMetric) get() float64 {
	fval := math.Float64frombits(atomic.LoadUint64(&c.valBits))
	ival := atomic.LoadUint64(&c.valInt)
	return fval + float64(ival)
}

func (c *counterMetric) Read(ts time.Time) *Counter {
	return &Counter{
		Timestamp: ts,
		Name:      c.name,
		Value:     c.get(),
	}
}

type CounterSeries []*Counter

// Rate calculates rate per second
func (s CounterSeries) Rate(start, end time.Time) float64 {
	if len(s) < 2 {
		return 0
	}
	r := s[len(s)-1].Value - s[0].Value
	prev := s[0].Value
	for _, p := range s {
		if p.Value < prev {
			r += prev
		}
		prev = p.Value
	}

	durationToStart := s[0].Timestamp.Sub(start).Seconds()
	durationToEnd := end.Sub(s[len(s)-1].Timestamp).Seconds()

	sampledInterval := s[len(s)-1].Timestamp.Sub(s[0].Timestamp).Seconds()
	averageDurationBetweenSamples := sampledInterval / float64(len(s))
	durationToZero := sampledInterval * (s[0].Value / r)

	if durationToZero < durationToStart {
		durationToStart = durationToZero
	}
	extrapolationThreshold := averageDurationBetweenSamples * 1.1
	extrapolateToInterval := sampledInterval

	if durationToStart < extrapolationThreshold {
		extrapolateToInterval += durationToStart
	} else {
		extrapolateToInterval += averageDurationBetweenSamples / 2
	}
	if durationToEnd < extrapolationThreshold {
		extrapolateToInterval += durationToEnd
	} else {
		extrapolateToInterval += averageDurationBetweenSamples / 2
	}
	factor := extrapolateToInterval / sampledInterval
	return r * factor
}
