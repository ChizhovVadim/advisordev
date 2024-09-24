package advisors

import (
	"advisordev/internal/domain"
)

type ConstVolatilityIndicator struct {
	stdVolatility float64
}

func (ind *ConstVolatilityIndicator) Add(candle domain.Candle) {}

func (ind *ConstVolatilityIndicator) Value() float64 {
	return ind.stdVolatility
}
