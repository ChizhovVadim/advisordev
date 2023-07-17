package advisors

import (
	"advisordev/internal/core"
	"log"
	"math"
)

func TestAdvisor(logger *log.Logger) core.Advisor {
	const stdVolatility = 0.006
	var advisor = MovingAverageAdvisor(stdVolatility)
	advisor = ApplyCandleValidation(advisor, logger)
	return advisor
}

func MovingAverageAdvisor(stdVolatility float64) core.Advisor {
	return GeneralAdvisorNew(
		50,
		stdVolatility,
		&MainSessionRebalanceInd{},
		&MovingAverageIng{period: 100},
		&VolatilityInd{standardVol: stdVolatility, period: 100},
	)
}

func GeneralAdvisorNew(
	emaPeriod int,
	stdVolatility float64,
	rebalanceInd IUpdateBoolIndicator,
	trendInd IUpdateFloat64Indicator,
	volatilityInd IUpdateFloat64Indicator,
) core.Advisor {
	type Details struct {
		Name     string
		Trend    float64
		VolRatio float64
	}
	var ratio float64
	var volRatio float64
	var details interface{}
	return func(candle core.Candle) core.Advice {
		rebalanceInd.Add(candle)
		trendInd.Add(candle)
		volatilityInd.Add(candle)
		if rebalanceInd.Value() {
			volRatio = math.Min(1, stdVolatility/volatilityInd.Value())
			var newRatio = trendInd.Value()
			ratio += (newRatio - ratio) / float64(emaPeriod)
			details = Details{
				Name:     "GeneralAdvisorNew",
				Trend:    newRatio,
				VolRatio: volRatio,
			}
		}
		return core.Advice{
			SecurityCode: candle.SecurityCode,
			DateTime:     candle.DateTime,
			Price:        candle.ClosePrice,
			Position:     ratio * volRatio,
			Details:      details,
		}
	}
}
