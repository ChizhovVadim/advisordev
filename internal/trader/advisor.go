package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/domain"
	"advisordev/internal/quik"
	"log/slog"
)

func initAdvisor(
	logger *slog.Logger,
	quikService *quik.QuikService,
	strategyConfig advisors.StrategyConfig,
	secInfo SecurityInfo,
	outAdvisor *domain.Advisor,
	outLastAdvice *domain.Advice,
) error {
	const interval = quik.CandleIntervalM5

	var advisor = advisors.Maindvisor(logger, strategyConfig)
	lastQuikCandles, err := quikService.GetLastCandles(secInfo.ClassCode, secInfo.Code, interval, 0)
	if err != nil {
		return err
	}

	var lastCandles []domain.Candle
	for _, quikCandle := range lastQuikCandles {
		var candle = convertQuikCandle(secInfo.Name, quikCandle)
		//if !candle.DateTime.Before(skipBefore) {
		lastCandles = append(lastCandles, candle)
		//}
	}

	// последний бар за сегодня может быть не завершен
	if len(lastCandles) > 0 && isToday(lastCandles[len(lastCandles)-1].DateTime) {
		lastCandles = lastCandles[:len(lastCandles)-1]
	}

	if len(lastCandles) == 0 {
		logger.Warn("Ready candles empty")
	} else {
		logger.Info("Ready candles",
			"First", lastCandles[0],
			"Last", lastCandles[len(lastCandles)-1],
			"Size", len(lastCandles))
	}
	var initAdvice domain.Advice
	for _, candle := range lastCandles {
		var advice = advisor(candle)
		if !advice.DateTime.IsZero() {
			initAdvice = advice
		}
	}
	logger.Info("Init advice",
		"advice", initAdvice)

	*outAdvisor = advisor
	*outLastAdvice = initAdvice
	return nil
}
