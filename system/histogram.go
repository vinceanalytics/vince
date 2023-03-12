// The MIT License (MIT)
//
// Copyright (c) 2019 VictoriaMetrics
// Copyright (c) 2023 VINCE
package system

import (
	"math"
	"sync"
	"time"
)

const (
	e10Min              = -9
	e10Max              = 18
	bucketsPerDecimal   = 18
	decimalBucketsCount = e10Max - e10Min
	bucketsCount        = decimalBucketsCount * bucketsPerDecimal
)

var bucketMultiplier = math.Pow(10, 1.0/bucketsPerDecimal)

// Histogram is a histogram for non-negative values with automatically created buckets.
//
// See https://medium.com/@valyala/improving-histogram-usability-for-prometheus-and-grafana-bc7e5df0e350
//
// Each bucket contains a counter for values in the given range.
// Each non-empty bucket is exposed via the following metric:
//
//	<metric_name>_bucket{<optional_tags>,vmrange="<start>...<end>"} <counter>
//
// Where:
//
//   - <metric_name> is the metric name passed to NewHistogram
//   - <optional_tags> is optional tags for the <metric_name>, which are passed to NewHistogram
//   - <start> and <end> - start and end values for the given bucket
//   - <counter> - the number of hits to the given bucket during Update* calls
//
// Histogram buckets can be converted to Prometheus-like buckets with `le` labels
// with `prometheus_buckets(<metric_name>_bucket)` function from PromQL extensions in VictoriaMetrics.
// (see https://github.com/VictoriaMetrics/VictoriaMetrics/wiki/MetricsQL ):
//
//	prometheus_buckets(request_duration_bucket)
//
// Time series produced by the Histogram have better compression ratio comparing to
// Prometheus histogram buckets with `le` labels, since they don't include counters
// for all the previous buckets.
//
// Zero histogram is usable.
type Histogram struct {
	name string
	mu   sync.Mutex

	decimalBuckets [decimalBucketsCount]*[bucketsPerDecimal]uint64

	lower uint64
	upper uint64

	sum float64
}

type History struct {
	Timestamp  time.Time `parquet:"timestamp,dict,zstd"`
	Name       string    `parquet:"name,dict,zstd"`
	Buckets    []Bucket
	Sum        float64 `parquet:"sum,dict,zstd"`
	TotalCount uint64  `parquet:"start,dict,zstd"`
}

func (h *History) Reset() {
	h.Buckets = h.Buckets[:0]
	h.Sum = 0
	h.TotalCount = 0
}

type Bucket struct {
	Start float64 `parquet:"start,dict,zstd"`
	End   float64 `parquet:"end,dict,zstd"`
	Count uint64  `parquet:"count,dict,zstd"`
}

// Reset resets the given histogram.
func (h *Histogram) Reset() {
	h.mu.Lock()
	for _, db := range h.decimalBuckets[:] {
		if db == nil {
			continue
		}
		for i := range db[:] {
			db[i] = 0
		}
	}
	h.lower = 0
	h.upper = 0
	h.sum = 0
	h.mu.Unlock()
}

// Update updates h with v.
//
// Negative values and NaNs are ignored.
func (h *Histogram) Update(v float64) {
	if math.IsNaN(v) || v < 0 {
		// Skip NaNs and negative values.
		return
	}
	bucketIdx := (math.Log10(v) - e10Min) * bucketsPerDecimal
	h.mu.Lock()
	h.sum += v
	if bucketIdx < 0 {
		h.lower++
	} else if bucketIdx >= bucketsCount {
		h.upper++
	} else {
		idx := uint(bucketIdx)
		if bucketIdx == float64(idx) && idx > 0 {
			// Edge case for 10^n values, which must go to the lower bucket
			// according to Prometheus logic for `le`-based histograms.
			idx--
		}
		decimalBucketIdx := idx / bucketsPerDecimal
		offset := idx % bucketsPerDecimal
		db := h.decimalBuckets[decimalBucketIdx]
		if db == nil {
			var b [bucketsPerDecimal]uint64
			db = &b
			h.decimalBuckets[decimalBucketIdx] = db
		}
		db[offset]++
	}
	h.mu.Unlock()
}

func (h *Histogram) Read(o *History) {
	h.mu.Lock()
	o.Name = h.name
	if h.lower > 0 {
		o.Buckets = append(o.Buckets, Bucket{
			Start: lowerBucketRange.Start,
			End:   lowerBucketRange.End,
			Count: h.lower,
		})
		o.TotalCount += h.lower
	}
	for decimalBucketIdx, db := range h.decimalBuckets[:] {
		if db == nil {
			continue
		}
		for offset, count := range db[:] {
			if count > 0 {
				bucketIdx := decimalBucketIdx*bucketsPerDecimal + offset
				vmrange := getVMRange(bucketIdx)
				o.Buckets = append(o.Buckets, Bucket{
					Start: vmrange.Start,
					End:   vmrange.End,
					Count: count,
				})
				o.TotalCount += count
			}
		}
	}
	if h.upper > 0 {
		o.Buckets = append(o.Buckets, Bucket{
			Start: upperBucketRange.Start,
			End:   upperBucketRange.End,
			Count: h.upper,
		})
		o.TotalCount += h.upper
	}
	o.Sum = h.sum
	h.mu.Unlock()
}

// UpdateDuration updates request duration based on the given startTime.
func (h *Histogram) UpdateDuration(startTime time.Time) {
	d := time.Since(startTime).Seconds()
	h.Update(d)
}

func getVMRange(bucketIdx int) Range {
	bucketRangesOnce.Do(initBucketRanges)
	return bucketRanges[bucketIdx]
}

func initBucketRanges() {
	v := math.Pow10(e10Min)
	start := v
	for i := 0; i < bucketsCount; i++ {
		v *= bucketMultiplier
		end := v
		bucketRanges[i] = Range{
			Start: start,
			End:   end,
		}
		start = end
	}
}

type Range struct {
	Start float64
	End   float64
}

var (
	lowerBucketRange = Range{
		End: math.Pow10(e10Min),
	}
	upperBucketRange = Range{
		Start: math.Pow10(e10Max),
		End:   math.Inf(0),
	}

	bucketRanges     [bucketsCount]Range
	bucketRangesOnce sync.Once
)
