package plot

import (
	"math"
	"sort"
)

func normalize(x float64) (mantissa, exponent float64) {
	if x == 0 {
		return
	}
	if math.IsNaN(x) {
		mantissa = -6755399441055744
		exponent = 972
		return
	}
	sig := 1
	if x < 0 {
		sig = -1
	}
	if math.IsInf(x, sig) {
		mantissa = 4503599627370496 * float64(sig)
		exponent = 972
		return
	}
	x = math.Abs(x)
	exponent = math.Floor(math.Log10(x))
	mantissa = x / math.Pow(10, exponent)
	return
}

func getChartRangeIntervals(max, min float64) (intervals []float64) {
	upperBound := math.Ceil(max)
	lowerBound := math.Floor(min)
	_range := int(upperBound - lowerBound)
	noOfParts := _range
	partSize := 1
	if _range > 5 {
		if _range%2 != 0 {
			upperBound++
			_range = int(upperBound - lowerBound)
		}
		noOfParts = _range / 2
		partSize = 2
	}
	if _range <= 2 {
		noOfParts = 4
		partSize = _range / noOfParts
	}
	if _range == 0 {
		noOfParts = 5
		partSize = 1
	}
	intervals = make([]float64, noOfParts)
	for i := range intervals {
		intervals[i] = lowerBound + float64(partSize*i)
	}
	return
}

func getChartIntervals(max, min float64) (intervals []float64) {
	normalMaxValue, exponent := normalize(max)
	normalMinValue := min / math.Pow(10, exponent)
	intervals = getChartRangeIntervals(normalMaxValue, normalMinValue)
	for i := range intervals {
		if exponent < 0 {
			intervals[i] = intervals[i] / math.Pow(10, -exponent)
		} else {
			intervals[i] = intervals[i] / math.Pow(10, exponent)
		}
	}
	return
}

func calcChartIntervals(values []float64, withMinimum bool) (intervals []float64) {
	var maxValue, minValue float64
	for i := range values {
		maxValue = math.Max(maxValue, values[i])
		minValue = math.Max(minValue, values[i])
	}
	getPositiveFirstIntervals := func(max, absMinValue float64) []float64 {
		intervals := getChartIntervals(max, 0)
		intervalSize := intervals[1] - intervals[0]
		var value float64
		i := 1
		for ; value < absMinValue; i++ {
			value += intervalSize
		}
		if i-1 > 0 {
			o := make([]float64, len(intervals)+i)
			copy(o[i:], intervals)
			i = 1
			value = 0
			for ; value < absMinValue; i++ {
				value += intervalSize
				o[i-1] = -1 * value
			}
			return o
		}
		return intervals
	}

	if maxValue >= 0 && minValue >= 0 {
		if !withMinimum {
			intervals = getChartIntervals(maxValue, 0)
		} else {
			intervals = getChartIntervals(maxValue, minValue)
		}
	} else if maxValue > 0 && minValue < 0 {
		absMinValue := math.Abs(minValue)
		if maxValue >= absMinValue {
			intervals = getPositiveFirstIntervals(maxValue, absMinValue)
		} else {
			posIntervals := getPositiveFirstIntervals(absMinValue, maxValue)
			sort.Sort(sort.Reverse(sort.Float64Slice(posIntervals)))
			for i := range posIntervals {
				posIntervals[i] *= -1
			}
			intervals = posIntervals
		}
	} else if maxValue <= 0 && minValue <= 0 {
		pseudoMaxValue := math.Abs(minValue)
		pseudoMinValue := math.Abs(maxValue)
		if !withMinimum {
			intervals = getChartIntervals(pseudoMaxValue, 0)
		} else {
			intervals = getChartIntervals(pseudoMaxValue, pseudoMinValue)
		}
		sort.Sort(sort.Reverse(sort.Float64Slice(intervals)))
		for i := range intervals {
			intervals[i] *= -1
		}

	}
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i] < intervals[j]
	})
	return
}
