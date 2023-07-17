package advisors

import "advisordev/internal/core"

func compare(a, b float64) float64 {
	if b > a {
		return 1
	}
	if b < a {
		return -1
	}
	return 0
}

func sumFloat64(source []float64) float64 {
	var result = 0.0
	for _, x := range source {
		result += x
	}
	return result
}

func mean(source []float64) float64 {
	return sumFloat64(source) / float64(len(source))
}

func isMainSessionOneDayCandles(x, y core.Candle) bool {
	return !x.DateTime.IsZero() &&
		!y.DateTime.IsZero() &&
		core.IsMainFortsSession(x.DateTime) &&
		core.IsMainFortsSession(y.DateTime) &&
		!core.IsNewDayStarted(x.DateTime, y.DateTime)
}
