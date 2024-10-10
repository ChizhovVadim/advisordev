package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/domain"
	"advisordev/internal/quik"
	"log/slog"
)

// Не совершает сделок на рынке, а только отслеживает инструмент и строит прогноз
type QuietStrategy struct {
	logger     *slog.Logger
	secInfo    SecurityInfo
	advisor    domain.Advisor
	lastAdvice domain.Advice
}

func initQuietStrategy(
	logger *slog.Logger,
	quikService *quik.QuikService,
	strategyConfig advisors.StrategyConfig,
) (*QuietStrategy, error) {
	logger = logger.
		With("security", strategyConfig.SecurityCode).
		With("advisor", strategyConfig.Name).
		With("strategy", "QuietStrategy")

	const interval = quik.CandleIntervalM5

	secInfo, err := getSecurityInfoHardCode(strategyConfig.SecurityCode)
	if err != nil {
		return nil, err
	}

	var advisor domain.Advisor
	var initAdvice domain.Advice
	err = initAdvisor(logger, quikService, strategyConfig, secInfo, &advisor, &initAdvice)
	if err != nil {
		return nil, err
	}

	isSubscribed, err := quikService.IsCandleSubscribed(secInfo.ClassCode, secInfo.Code, interval)
	if err != nil {
		return nil, err
	}
	if !isSubscribed {
		err = quikService.SubscribeCandles(secInfo.ClassCode, secInfo.Code, interval)
		if err != nil {
			return nil, err
		}
		logger.Info("Subscribed")
	}

	return &QuietStrategy{
		logger:     logger,
		secInfo:    secInfo,
		advisor:    advisor,
		lastAdvice: initAdvice,
	}, nil
}

func (s *QuietStrategy) OnNewCandle(
	newCandle quik.Candle,
) (bool, error) {
	if !(newCandle.Interval == quik.CandleIntervalM5 &&
		newCandle.SecCode == s.secInfo.Code) {
		return false, nil
	}

	var myCandle = convertQuikCandle(s.secInfo.Name, newCandle)
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
