package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/domain"
	"advisordev/internal/quik"
	"advisordev/internal/utils"
	"log/slog"
)

// Не совершает сделок на рынке, а только отслеживает инструмент и строит прогноз
type QuietStrategy struct {
	logger       *slog.Logger
	securityName string
	securityCode string
	advisor      domain.Advisor
	lastAdvice   domain.Advice
}

func initQuietStrategy(
	logger *slog.Logger,
	quikService *quik.QuikService,
	strategyConfig advisors.StrategyConfig,
) (*QuietStrategy, error) {
	const interval = quik.CandleIntervalM5

	var advisor = advisors.Maindvisor(logger, strategyConfig)
	var securityName = strategyConfig.SecurityCode

	var classCode = strategyConfig.ClassCode
	if classCode == "" {
		classCode = FuturesClassCode
	}

	var securityCode string
	var err error
	if classCode == FuturesClassCode {
		securityCode, err = utils.EncodeSecurity(securityName)
		if err != nil {
			return nil, err
		}
	} else {
		securityCode = securityName
	}

	lastQuikCandles, err := quikService.GetLastCandles(classCode, securityCode, interval, 0)
	if err != nil {
		return nil, err
	}

	var lastCandles []domain.Candle
	for _, quikCandle := range lastQuikCandles {
		var candle = convertQuikCandle(securityName, quikCandle)
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

	err = quikService.SubscribeCandles(classCode, securityCode, interval)
	if err != nil {
		return nil, err
	}
	return &QuietStrategy{
		logger:       logger,
		securityName: securityName,
		securityCode: securityCode,
		advisor:      advisor,
		lastAdvice:   initAdvice,
	}, nil
}

func (s *QuietStrategy) OnNewCandle(
	newCandle quik.Candle,
) (bool, error) {
	if !(newCandle.Interval == quik.CandleIntervalM5 &&
		s.securityCode == newCandle.SecCode) {
		return false, nil
	}

	var myCandle = convertQuikCandle(s.securityName, newCandle)
	var advice = s.advisor(myCandle)
	if advice.DateTime.IsZero() {
		return false, nil
	}
	if advice.Position != s.lastAdvice.Position {
		s.logger.Info("New advice",
			"Advice", advice)
	}
	s.lastAdvice = advice
	return false, nil
}

func (s *QuietStrategy) CheckPosition() error {
	return nil
}
