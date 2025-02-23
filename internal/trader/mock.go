package trader

import (
	"advisordev/internal/domain"
	"log/slog"
)

type MockTrader struct {
	logger    *slog.Logger
	positions map[string]float64
}

func NewMockTrader(
	logger *slog.Logger,
) *MockTrader {
	return &MockTrader{
		logger:    logger,
		positions: make(map[string]float64),
	}
}

func (c *MockTrader) IsConnected() (bool, error) {
	return true, nil
}

func (c *MockTrader) IncomingAmount(portfolio domain.PortfolioInfo) (float64, error) {
	return 1_000_000, nil
}

func (c *MockTrader) GetPosition(portfolio domain.PortfolioInfo, security domain.SecurityInfo) (float64, error) {
	return c.positions[security.Code], nil
}

func (c *MockTrader) RegisterOrder(order domain.Order) error {
	c.positions[order.Security.Code] += float64(order.Volume)
	return nil
}
