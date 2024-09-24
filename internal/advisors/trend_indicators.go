package advisors

import (
	"advisordev/internal/domain"
)

func compare(a, b float64) float64 {
	if b > a {
		return 1
	}
	if b < a {
		return -1
	}
	return 0
}

// только для примера
type EMATrendIndicator struct {
	period int
	size   int
	ema    float64
	last   domain.Candle
}

func (ind *EMATrendIndicator) Add(candle domain.Candle) {
	if ind.size == 0 {
		ind.ema = candle.ClosePrice
	} else {
		ind.ema += (candle.ClosePrice - ind.ema) / float64(ind.period)
	}
	ind.last = candle
	ind.size += 1
}

func (ind *EMATrendIndicator) Value() float64 {
	if ind.size < ind.period {
		return 0
	}
	return compare(ind.ema, ind.last.ClosePrice)
}
