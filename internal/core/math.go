package core

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
		panic("Sequence contains no elements")
	}

	stDev = math.Sqrt(M2 / float64(n))
	return
}

func StDev(source []float64) float64 {
	var _, stDev = Moments(source)
	return stDev
}
