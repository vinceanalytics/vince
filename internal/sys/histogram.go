package sys

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

type Histogram struct {
	// Mu gurantees synchronous update for all the counters and sum.
	//
	// Do not use sync.RWMutex, since it has zero sense from performance PoV.
	// It only complicates the code.
	mu sync.Mutex

	// decimalBuckets contains counters for histogram buckets
	decimalBuckets [decimalBucketsCount]*[bucketsPerDecimal]uint64

	// lower is the number of values, which hit the lower bucket
	lower uint64

	// upper is the number of values, which hit the upper bucket
	upper uint64

	// sum is the sum of all the values put into Histogram
	sum float64
}

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

func (h *Histogram) VisitNonZeroBuckets(f func(bucket int, count uint64)) {
	h.mu.Lock()
	if h.lower > 0 {
		f(lowerBucketRange, h.lower)
	}
	for decimalBucketIdx, db := range h.decimalBuckets[:] {
		if db == nil {
			continue
		}
		for offset, count := range db[:] {
			if count > 0 {
				bucketIdx := decimalBucketIdx*bucketsPerDecimal + offset
				f(bucketIdx, count)
			}
		}
	}
	if h.upper > 0 {
		f(upperBucketRange, h.upper)
	}
	h.mu.Unlock()
}

func (h *Histogram) UpdateDuration(startTime time.Time) {
	d := time.Since(startTime).Seconds()
	h.Update(d)
}

func (h *Histogram) Marshal(f func(bucket int, count uint64)) (total uint64, sum float64) {
	h.VisitNonZeroBuckets(func(bucket int, count uint64) {
		total += count
		f(bucket, count)
	})
	h.mu.Lock()
	sum = h.sum
	h.mu.Unlock()
	return
}

var startBucket = math.Pow10(e10Min)

func bucket(idx int) (start, end float64) {
	if idx == 0 {
		return startBucket, bucketRanges[idx]
	}
	return bucketRanges[idx-1], bucketRanges[idx]
}

func init() {
	v := math.Pow10(e10Min)
	for i := 0; i < bucketsCount; i++ {
		v *= bucketMultiplier
		bucketRanges[i] = v
	}
}

var (
	lowerBucketRange = 0
	upperBucketRange = bucketsCount
	bucketRanges     [bucketsCount]float64
)
