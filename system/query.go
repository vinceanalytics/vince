package system

import "time"

// returns rate per second on counter.
func rate(ts []int64, values []float64) float64 {
	if len(values) != len(ts) {
		return 0
	}
	if len(values) < 2 {
		return 0
	}
	dv := values[len(values)-1] - values[0]
	dt := ts[len(ts)-1] - ts[0]
	return dv / (float64(dt) / 1e3)
}

func sum(ts []int64, values []float64) (o float64) {
	for _, v := range values {
		o += v
	}
	return
}

func roll(shared, ts []int64, values []float64, window int64,
	f func([]int64, []float64) float64) (ov []float64) {
	i := 0
	j := 0
	ni := 0
	nj := 0
	ov = make([]float64, 0, len(shared))
	for _, tEnd := range shared {
		tStart := tEnd - window
		ni = seekFirstTimestampIdxAfter(ts[i:], tStart, ni)
		i += ni
		if j < i {
			j = i
		}
		nj = seekFirstTimestampIdxAfter(ts[j:], tEnd, nj)
		j += nj
		p := f(ts[i:j], values[i:j])
		ov = append(ov, p)
	}
	return
}

// seekFirstTimestampIdxAfter and binarySearchInt64 are copy pasted  from VictoriaMetrics
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

var step = (5 * time.Minute).Milliseconds()

func sharedTimestamp(start, end int64) []int64 {
	points := 1 + (end-start)/step
	timestamps := make([]int64, points)
	for i := range timestamps {
		timestamps[i] = start
		start += step
	}
	return timestamps
}
