package trader

import (
	"advisordev/internal/domain"
	"fmt"
	"log/slog"
	"time"
)

type StrategyService struct {
	logger    *slog.Logger
	trader    domain.ITrader
	portfolio domain.PortfolioInfo
	security  domain.SecurityInfo
	advisor   string
	amount    float64
	position  int
	basePrice float64
}

func initStrategy(
	logger *slog.Logger,
	config StrategyConfig,
	trader domain.ITrader,
	portfolio domain.PortfolioInfo,
	security domain.SecurityInfo,
	amount float64,
) (*StrategyService, error) {
	logger = logger.With(
		//"portfolio", portfolio.Portfolio,
		"advisor", config.Advisor,
		"security", security.Name)

	pos, err := trader.GetPosition(portfolio, security)
	if err != nil {
		return nil, err
	}
	var initPosition = int(pos)
	logger.Info("Init position",
		"Position", initPosition)

	if config.Weight != 0 {
		amount *= config.Weight
	}

	return &StrategyService{
		logger:    logger,
		trader:    trader,
		portfolio: portfolio,
		security:  security,
		advisor:   config.Advisor,
		amount:    amount,
		position:  initPosition,
	}, nil
}

func (strategy *StrategyService) OnSignal(advice domain.Advice, outOrderRegistered *bool) error {
	// стратегия следит только за своими сигналами
	if !(advice.SecurityCode == strategy.security.Code &&
		advice.Advisor == strategy.advisor) {
		return nil
	}

	// считаем, что сигнал слишком старый
	if time.Since(advice.DateTime) >= 9*time.Minute {
		return nil
	}

	if strategy.basePrice == 0 {
		strategy.basePrice = advice.Price
		strategy.logger.Info("Init base price",
			"DateTime", advice.DateTime,
			"Price", advice.Price)
	}

	var position = strategy.amount / (strategy.basePrice * strategy.security.Lever) * advice.Position
	var volume = int(position - float64(strategy.position))
	if volume == 0 {
		return nil
	}
	strategy.logger.Info("New advice",
		"Advice", advice)
	if !strategy.CheckPosition() {
		return nil
	}
	var price = priceWithSlippage(advice.Price, volume)
	strategy.logger.Info("Register order",
		"Price", price,
		"Volume", volume)
	var err = strategy.trader.RegisterOrder(domain.Order{
		Portfolio: strategy.portfolio,
		Security:  strategy.security,
		Volume:    volume,
		Price:     price,
	})
	if err != nil {
		return fmt.Errorf("RegisterOrder failed %w", err)
	}
	*outOrderRegistered = true
	strategy.position += volume
	return nil
}

func (strategy *StrategyService) CheckPosition() bool {
	pos, err := strategy.trader.GetPosition(strategy.portfolio, strategy.security)
	if err != nil {
		return false
	}
	var traderPosition = int(pos)
	if strategy.position == traderPosition {
		strategy.logger.Info("Check position",
			"Position", strategy.position,
			"Status", "+")
		return true
	} else {
		strategy.logger.Warn("Check position",
			"StrategyPosition", strategy.position,
			"TraderPosition", traderPosition,
			"Status", "!")
		return false
	}
}

func priceWithSlippage(price float64, volume int) float64 {
	const Slippage = 0.001
	if volume > 0 {
		return price * (1 + Slippage)
	} else {
		return price * (1 - Slippage)
	}
}
