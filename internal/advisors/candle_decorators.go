package advisors

import (
	"advisordev/internal/domain"
)

func newEmptyCandleDecorator() func(domain.Candle) (domain.Candle, bool) {
	return func(c domain.Candle) (domain.Candle, bool) {
		return c, true
	}
}
