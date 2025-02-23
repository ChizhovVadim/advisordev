package trader

import (
	"advisordev/internal/domain"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"
)

func executeSignals(
	ctx context.Context,
	logger *slog.Logger,
	client Client,
	securityNames []string,
	securityInformator domain.ISecurityInformator,
	connector domain.ITrader,
	advices <-chan domain.Advice,
) error {
	logger = logger.With("name", "executeSignals")
	connected, err := connector.IsConnected()
	if err != nil {
		return err
	}
	if !connected {
		return errors.New("quik is not connected")
	}

	var portfolio = domain.PortfolioInfo{
		Firm:      client.Firm,
		Portfolio: client.Portfolio,
	}
	startAmount, err := connector.IncomingAmount(portfolio)
	if err != nil {
		return err
	}
	availableAmount := calcAvailableAmount(startAmount, client)
	logger.Info("Init portfolio",
		"Amount", startAmount,
		"AvailableAmount", availableAmount)
	if availableAmount == 0 {
		logger.Warn("availableAmount zero")
	}

	var strategyList []Strategy
	for _, securityName := range securityNames {
		var strategy, err = initStrategy(
			logger, connector, portfolio, availableAmount, securityName, securityInformator)
		if err != nil {
			return err
		}
		strategyList = append(strategyList, strategy)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case advice, ok := <-advices:
			if !ok {
				return nil
			}
			for i := range strategyList {
				var err = strategyList[i].OnNewAdvice(advice)
				if err != nil {
					logger.Error("OnNewAdvice failed", "error", err)
				}
			}
		}
	}
}

type Strategy struct {
	logger    *slog.Logger
	trader    domain.ITrader
	portfolio domain.PortfolioInfo
	amount    float64
	security  domain.SecurityInfo
	position  int
	basePrice float64
}

func initStrategy(
	logger *slog.Logger,
	trader domain.ITrader,
	portfolio domain.PortfolioInfo,
	amount float64,
	securityName string,
	securityInformator domain.ISecurityInformator,
) (Strategy, error) {
	logger = logger.With("security", securityName)

	security, err := securityInformator.GetSecurityInfo(securityName)
	if err != nil {
		return Strategy{}, err
	}
	pos, err := trader.GetPosition(portfolio, security)
	if err != nil {
		return Strategy{}, err
	}
	var initPosition = int(pos)
	logger.Info("Init position",
		"Position", initPosition)

	return Strategy{
		logger:    logger,
		trader:    trader,
		portfolio: portfolio,
		amount:    amount,
		security:  security,
		position:  initPosition,
	}, nil
}

func (strategy *Strategy) OnNewAdvice(advice domain.Advice) error {
	if strategy.security.Code != advice.SecurityCode ||
		time.Since(advice.DateTime) >= 9*time.Minute /*считаем, что сигнал слишком старый*/ {
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
	strategy.position += volume
	return nil
}

func (strategy *Strategy) CheckPosition() bool {
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

func calcAvailableAmount(startAmount float64, client Client) float64 {
	var result float64
	if client.Amount > 0 {
		result = client.Amount
	} else {
		result = startAmount
	}
	if client.MaxAmount > 0 {
		result = math.Min(result, client.MaxAmount)
	}
	if 0 < client.Weight && client.Weight < 1 {
		result *= client.Weight
	}
	return result
}

func priceWithSlippage(price float64, volume int) float64 {
	const Slippage = 0.001
	if volume > 0 {
		return price * (1 + Slippage)
	} else {
		return price * (1 - Slippage)
	}
}
