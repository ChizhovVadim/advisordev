package advisors

import (
	"advisordev/internal/domain"
	"advisordev/internal/utils"
	"time"
)

func newTimeRebalance4() *timeRebalanceIndicator {
	return &timeRebalanceIndicator{periods: []time.Duration{
		time.Duration(11*time.Hour + 30*time.Minute),
		time.Duration(13*time.Hour + 30*time.Minute),
		time.Duration(15*time.Hour + 30*time.Minute),
		time.Duration(17*time.Hour + 30*time.Minute),
	}}
}

type timeRebalanceIndicator struct {
	periods    []time.Duration
	lastCandle domain.Candle
	value      bool
}

func (i *timeRebalanceIndicator) Add(candle domain.Candle) {
	i.value = false
	if !utils.IsMainFortsSession(candle.DateTime) {
		return
	}
	if !i.lastCandle.DateTime.IsZero() {
		for _, period := range i.periods {
			if utils.IsNewPeriod(i.lastCandle.DateTime, candle.DateTime, period) {
				i.value = true
				break
			}
		}
	}
	i.lastCandle = candle
}

func (i *timeRebalanceIndicator) Value() bool {
	return i.value
}
