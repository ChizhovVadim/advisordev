package history

import (
	"advisordev/internal/domain"
	"iter"
	"math"
	"time"
)

func SingleContractHprs(
	candles iter.Seq2[domain.Candle, error],
	advisor domain.Advisor,
	slippage float64,
	skipPnl func(time.Time, time.Time) bool) ([]DateSum, error) {

	var result []DateSum
	var pnl = 0.0
	var baseAdvice = domain.Advice{}
	var lastAdvice = domain.Advice{}

	for candle, err := range candles {
		if err != nil {
			return nil, err
		}
		var advice = advisor(candle)
		if advice.DateTime.IsZero() {
			continue
		}
		if baseAdvice.DateTime.IsZero() {
			baseAdvice = advice
			lastAdvice = advice
			continue
		}
		if isNewFortsDateStarted(lastAdvice.DateTime, advice.DateTime) {
			var ds = DateSum{Date: dateTimeToDate(lastAdvice.DateTime), Sum: 1 + pnl/baseAdvice.Price}
			result = append(result, ds)
			pnl = 0
			baseAdvice = lastAdvice
		}
		if !skipPnl(lastAdvice.DateTime, advice.DateTime) {
			pnl += lastAdvice.Position*(advice.Price-lastAdvice.Price) -
				slippage*advice.Price*math.Abs(advice.Position-lastAdvice.Position)
		}
		lastAdvice = advice
	}

	if !lastAdvice.DateTime.IsZero() {
		var ds = DateSum{Date: dateTimeToDate(lastAdvice.DateTime), Sum: 1 + pnl/baseAdvice.Price}
		result = append(result, ds)
	}
	return result, nil
}
