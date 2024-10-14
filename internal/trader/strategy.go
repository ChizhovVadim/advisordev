package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/domain"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type Strategy struct {
	logger           *slog.Logger
	connector        domain.IConnector
	availableAmount  float64
	portfolio        domain.PortfolioInfo
	security         domain.SecurityInfo
	advisor          domain.Advisor
	lastAdvice       domain.Advice
	strategyPosition int
	basePrice        float64
}

func initStrategy(
	logger *slog.Logger,
	connector domain.IConnector,
	availableAmount float64,
	porftolio domain.PortfolioInfo,
	strategyConfig advisors.StrategyConfig,
) (*Strategy, error) {

	if availableAmount == 0 {
		return nil, errors.New("availableAmount zero")
	}

	logger = logger.
		With("security", strategyConfig.SecurityCode).
		With("advisor", strategyConfig.Name).
		With("strategy", "Strategy")

	security, err := getSecurityInfoHardCode(strategyConfig.SecurityCode)
	if err != nil {
		return nil, err
	}

	pos, err := connector.GetPosition(porftolio, security)
	if err != nil {
		return nil, err
	}
	var initPosition = int(pos)
	logger.Info("Init position",
		"Position", initPosition)

	var advisor domain.Advisor
	var initAdvice domain.Advice
	err = initAdvisor(logger, connector, strategyConfig, security, &advisor, &initAdvice)
	if err != nil {
		return nil, err
	}

	return &Strategy{
		logger:           logger,
		connector:        connector,
		availableAmount:  availableAmount,
		portfolio:        porftolio,
		security:         security,
		advisor:          advisor,
		lastAdvice:       initAdvice,
		strategyPosition: initPosition,
		basePrice:        0,
	}, nil
}

func (s *Strategy) OnNewCandle(
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
	s.lastAdvice = advice
	if time.Since(advice.DateTime) >= 9*time.Minute {
		return false, nil
	}
	if s.basePrice == 0 {
		s.basePrice = myCandle.ClosePrice
		s.logger.Info("Init base price",
			"Candle", myCandle)
	}
	// размер лота
	var position = s.availableAmount / (s.basePrice * s.security.Lever) * advice.Position
	var volume = int(position - float64(s.strategyPosition))
	if volume == 0 {
		return false, nil
	}
	s.logger.Info("New advice",
		"Advice", advice)
	err := s.CheckPosition()
	if err != nil {
		return false, err
	}
	var price = priceWithSlippage(advice.Price, volume)
	s.logger.Info("Register order",
		"Price", price,
		"Volume", volume)
	err = s.connector.RegisterOrder(s.portfolio, s.security, volume, price)
	if err != nil {
		return false, fmt.Errorf("registerOrder failed %w", err)
	}
	s.strategyPosition += volume
	return true, nil
}

func (s *Strategy) CheckPosition() error {
	pos, err := s.connector.GetPosition(s.portfolio, s.security)
	if err != nil {
		return err
	}
	var traderPosition = int(pos)
	if s.strategyPosition == traderPosition {
		s.logger.Info("Check position",
			"Position", s.strategyPosition,
			"Status", "+")
		return nil
	} else {
		s.logger.Warn("Check position",
			"StrategyPosition", s.strategyPosition,
			"TraderPosition", traderPosition,
			"Status", "!")
		return fmt.Errorf("StrategyPosition!=TraderPosition")
	}
}
