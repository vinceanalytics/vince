package sys

import "math"

var scale = math.Pow(10, float64(3))

func fromFloat(f float64) int64 {
	return int64(f * scale)
}
func toFloat(f int64) float64 {
	return float64(f) / scale
}
