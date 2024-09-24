package utils

import "math"

func Moments(source []float64) (mean, stDev float64) {
	var n = 0
	var M2 = 0.0

	for _, x := range source {
		n++
		var delta = x - mean
		mean += delta / float64(n)
		M2 += delta * (x - mean)
	}

	if n == 0 {
		return math.NaN(), math.NaN()
	}

	stDev = math.Sqrt(M2 / float64(n))
	return
}

func StDev(source []float64) float64 {
	var _, stDev = Moments(source)
	return stDev
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
