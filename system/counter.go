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
