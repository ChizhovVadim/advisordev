package domain

import (
	"iter"
)

type ICandleStorage interface {
	Candles(securityCode string) iter.Seq2[Candle, error]
}

type IAdvisorService interface {
	PublishAdvice(advice Advice) error
	GetLastAdvices() ([]Advice, error)
}

type ISecurityInformator interface {
	GetSecurityInfo(securityName string) (SecurityInfo, error)
}

type ITrader interface {
	//Close() error
	IsConnected() (bool, error)
	IncomingAmount(portfolio PortfolioInfo) (float64, error)
	GetPosition(portfolio PortfolioInfo, security SecurityInfo) (float64, error)
	RegisterOrder(order Order) error
	GetLastCandles(security SecurityInfo, timeframe string) ([]Candle, error)
	SubscribeCandles(security SecurityInfo, timeframe string) error
	//HandleCallbacks(ctx context.Context, candles chan<- Candle, candlesCallbackActive *atomic.Bool) error
}
