package advisors

import (
	"advisordev/internal/core"
	"fmt"
	"log"
	"math"
)

func ApplyCandleValidation(advisor core.Advisor, logger *log.Logger) core.Advisor {
	var lastCandle = core.Candle{}
	return func(candle core.Candle) core.Advice {
		if !lastCandle.DateTime.IsZero() && !candle.DateTime.After(lastCandle.DateTime) {
			logger.Printf("Invalid candle order %v %v", lastCandle, candle)
			return core.Advice{}
		}
		if !lastCandle.DateTime.IsZero() {
			var change = math.Log(candle.ClosePrice / lastCandle.ClosePrice)
			if math.Abs(change) >= 0.1 {
				logger.Printf("Big jump %v %v", lastCandle, candle)
			}
		}
		lastCandle = candle
		return advisor(candle)
	}
}

func CombineAdvisors(advisors []core.Advisor, weights []float64) core.Advisor {
	if weights == nil {
		weights = make([]float64, len(advisors))
		var w = 1.0 / float64(len(weights))
		for i := range weights {
			weights[i] = w
		}
	}
	if len(advisors) != len(weights) {
		panic(fmt.Errorf("len(advisors) != len(weights)"))
	}
	return func(candle core.Candle) core.Advice {
		var ratio = 0.0
		var hasValue = true
		var childs []core.Advice
		for i := range advisors {
			var advice = advisors[i](candle)
			if advice.DateTime.IsZero() {
				hasValue = false
			} else {
				ratio += weights[i] * advice.Position
			}
			childs = append(childs, advice)
		}
		if !hasValue {
			return core.Advice{}
		}
		return core.Advice{
			SecurityCode: candle.SecurityCode,
			DateTime:     candle.DateTime,
			Price:        candle.ClosePrice,
			Position:     ratio,
			Details:      childs}
	}
}
