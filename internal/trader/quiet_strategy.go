package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/domain"
	"log/slog"
)

// Не совершает сделок на рынке, а только отслеживает инструмент и строит прогноз
type QuietStrategy struct {
	logger     *slog.Logger
	security   domain.SecurityInfo
	advisor    domain.Advisor
	lastAdvice domain.Advice
}

func initQuietStrategy(
	logger *slog.Logger,
	connector domain.IConnector,
	strategyConfig advisors.StrategyConfig,
) (*QuietStrategy, error) {
	logger = logger.
		With("security", strategyConfig.SecurityCode).
		With("advisor", strategyConfig.Name).
		With("strategy", "QuietStrategy")

	security, err := getSecurityInfoHardCode(strategyConfig.SecurityCode)
	if err != nil {
		return nil, err
	}

	var advisor domain.Advisor
	var initAdvice domain.Advice
	err = initAdvisor(logger, connector, strategyConfig, security, &advisor, &initAdvice)
	if err != nil {
		return nil, err
	}

	return &QuietStrategy{
		logger:     logger,
		security:   security,
		advisor:    advisor,
		lastAdvice: initAdvice,
	}, nil
}

func (s *QuietStrategy) OnNewCandle(
	myCandle domain.Candle,
) (bool, error) {
	//TODO myCandle.Interval == candles.TFMinutes5
	if !(myCandle.SecurityCode == s.security.Code) {
		return false, nil
	}

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
