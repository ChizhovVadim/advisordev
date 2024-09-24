package advisors

import (
	"advisordev/internal/domain"
	"math"
)

func Sept24ExperimentAdvisor(
	stdVolatility float64,
) domain.Advisor {
	return generalAdvisor(
		"Sept24ExperimentAdvisor",
		stdVolatility,
		&ConstVolatilityIndicator{stdVolatility: stdVolatility},
		newEmptyCandleDecorator(),
		newTimeRebalance4(),
		2,
		&EMATrendIndicator{period: 150},
	)
}

func generalAdvisor(
	name string,
	stdVolatility float64,
	volInd IUpdateFloat64Indicator,
	candleDecorator func(domain.Candle) (domain.Candle, bool),
	rebalanceInd IUpdateBoolIndicator,
	emaPeriod int,
	trendInd IUpdateFloat64Indicator,
) domain.Advisor {
	type Details struct {
		Name     string
		Trend    float64
		VolRatio float64
	}

	var ratio float64
	var volRatio float64
	var details interface{}

	return func(candle domain.Candle) domain.Advice {
		volInd.Add(candle)
		{
			var candle, ok = candleDecorator(candle)
			if !ok {
				return domain.Advice{}
			}
			trendInd.Add(candle)
			rebalanceInd.Add(candle)
			if rebalanceInd.Value() {
				var volatility = volInd.Value()
				volRatio = math.Min(1, stdVolatility/volatility)
				var trend = trendInd.Value()
				ratio += (trend - ratio) / float64(emaPeriod)
				details = Details{
					Name:     name,
					Trend:    trend,
					VolRatio: volRatio,
				}
			}
		}
		return domain.Advice{
			SecurityCode: candle.SecurityCode,
			DateTime:     candle.DateTime,
			Price:        candle.ClosePrice,
			Position:     ratio * volRatio,
			Details:      details,
		}
	}
}
