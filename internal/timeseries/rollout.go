package timeseries

import (
	"math"

	"github.com/vinceanalytics/vince/pkg/log"
)

var (
	nan = math.NaN()
)

// This function was ported from VictoriaMetrics project.
func rollUp(window int64, values []float64, ts []int64, shared []int64, f func(*rollOptions) float64) (o []float64) {
	o = make([]float64, 0, len(shared))
	i := 0
	j := 0
	ni := 0
	nj := 0
	var r rollOptions
	for _, tEnd := range shared {
		tStart := tEnd - window
		ni = seekFirstTimestampIdxAfter(ts[i:], tStart, ni)
		i += ni
		if j < i {
			j = i
		}
		nj = seekFirstTimestampIdxAfter(ts[j:], tEnd, nj)
		j += nj

		r.prevValue = nan
		r.prevTimestamp = tStart
		if i < len(ts) && i > 0 && ts[i-1] > r.prevTimestamp {
			r.prevValue = values[i-1]
			r.prevTimestamp = ts[i-1]
		}
		r.values = values[i:j]
		r.ts = ts[i:j]
		r.realPrevValue = nan
		if i > 0 {
			r.realPrevValue = values[i-1]
		}
		r.realNextValue = nan
		if j < len(values) {
			r.realNextValue = values[j]
		}
		r.currTimestamp = tEnd
		value := f(&r)
		o = append(o, value)
	}
	return
}

type rollOptions struct {
	prevValue     float64
	prevTimestamp int64
	values        []float64
	ts            []int64
	realPrevValue float64
	realNextValue float64
	currTimestamp int64
}

func sharedTS(start, end, step int64) []int64 {
	if step <= 0 {
		log.Get().
			Panic().Int64("step", step).Msg("step > 0")
	}
	if start > end {
		log.Get().
			Panic().
			Int64("step", step).
			Int64("start", start).
			Int64("end", end).
			Msg("expect start < end")
	}
	points := 1 + (end-start)/step
	timestamps := make([]int64, points)
	for i := range timestamps {
		timestamps[i] = start
		start += step
	}
	return timestamps
}

// Copy Pasted from VictoriaMetrics
func seekFirstTimestampIdxAfter(timestamps []int64, seekTimestamp int64, nHint int) int {
	if len(timestamps) == 0 || timestamps[0] > seekTimestamp {
		return 0
	}
	startIdx := nHint - 2
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(timestamps) {
		startIdx = len(timestamps) - 1
	}
	endIdx := nHint + 2
	if endIdx > len(timestamps) {
		endIdx = len(timestamps)
	}
	if startIdx > 0 && timestamps[startIdx] <= seekTimestamp {
		timestamps = timestamps[startIdx:]
		endIdx -= startIdx
	} else {
		startIdx = 0
	}
	if endIdx < len(timestamps) && timestamps[endIdx] > seekTimestamp {
		timestamps = timestamps[:endIdx]
	}
	if len(timestamps) < 16 {
		// Fast path: the number of timestamps to search is small, so scan them all.
		for i, timestamp := range timestamps {
			if timestamp > seekTimestamp {
				return startIdx + i
			}
		}
		return startIdx + len(timestamps)
	}
	// Slow path: too big len(timestamps), so use binary search.
	i := binarySearchInt64(timestamps, seekTimestamp+1)
	return startIdx + int(i)
}

func binarySearchInt64(a []int64, v int64) uint {
	// Copy-pasted sort.Search from https://golang.org/src/sort/search.go?s=2246:2286#L49
	i, j := uint(0), uint(len(a))
	for i < j {
		h := (i + j) >> 1
		if h < uint(len(a)) && a[h] < v {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}
