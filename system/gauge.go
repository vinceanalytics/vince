package system

import (
	"math"
	"sync/atomic"
	"time"
)

type Gauge struct {
	Timestamp time.Time `parquet:"timestamp,dict,zstd"`
	Name      string    `parquet:"name,dict,zstd"`
	Value     float64   `parquet:"value,dict,zstd"`
}

type gaugeMetric struct {
	valBits uint64
	name    string
}

func (g *gaugeMetric) Set(val float64) {
	atomic.StoreUint64(&g.valBits, math.Float64bits(val))
}

func (g *gaugeMetric) Read(ts time.Time) *Gauge {
	return &Gauge{
		Timestamp: ts,
		Name:      g.name,
		Value:     math.Float64frombits(atomic.LoadUint64(&g.valBits)),
	}
}
