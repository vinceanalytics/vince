package system

import (
	"math"
	"sync/atomic"
)

type gaugeMetric struct {
	valBits uint64
	name    string
}

func (g *gaugeMetric) Set(val float64) {
	atomic.StoreUint64(&g.valBits, math.Float64bits(val))
}

func (g *gaugeMetric) get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.valBits))
}
