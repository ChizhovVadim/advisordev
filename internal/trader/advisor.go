package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/domain"
	"log/slog"
)

func initAdvisor(
	logger *slog.Logger,
	connector domain.IConnector,
	strategyConfig advisors.StrategyConfig,
	security domain.SecurityInfo,
	outAdvisor *domain.Advisor,
	outLastAdvice *domain.Advice,
) error {
	const timeframe = domain.CandleIntervalMinutes5

	lastCandles, err := connector.GetLastCandles(security, timeframe)
	if err != nil {
		return err
	}

	if len(lastCandles) == 0 {
		logger.Warn("Ready candles empty")
	} else {
		logger.Info("Ready candles",
			"First", lastCandles[0],
			"Last", lastCandles[len(lastCandles)-1],
			"Size", len(lastCandles))
	}

	var advisor = advisors.Maindvisor(logger, strategyConfig)
	var initAdvice domain.Advice
	for _, candle := range lastCandles {
		var advice = advisor(candle)
		if !advice.DateTime.IsZero() {
			initAdvice = advice
		}
	}
	logger.Info("Init advice",
		"advice", initAdvice)

	err = connector.SubscribeCandles(security, timeframe)
	if err != nil {
		return err
	}

	*outAdvisor = advisor
	*outLastAdvice = initAdvice
	return nil
}
