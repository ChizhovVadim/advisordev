package domain

const FuturesClassCode = "SPBFUT"

type PortfolioInfo struct {
	Firm      string
	Portfolio string
}

type SecurityInfo struct {
	// Название инструмента
	Name string
	// Код инструмента
	Code string
	// Код класса
	ClassCode string
	// точность (кол-во знаков после запятой). Если шаг цены может быть не круглым (0.05), то этого будет недостаточно.
	PricePrecision int
	// шаг цены
	PriceStep float64
	// Стоимость шага цены
	PriceStepCost float64
	// Плечо. Для фьючерсов = PriceStepCost/PriceStep.
	Lever float64
}

type IConnector interface {
	IsConnected() (bool, error)
	IncomingAmount(portfolio PortfolioInfo) (float64, error)
	GetPosition(portfolio PortfolioInfo, security SecurityInfo) (float64, error)
	RegisterOrder(portfolio PortfolioInfo, security SecurityInfo, volume int, price float64) error
	GetLastCandles(security SecurityInfo, timeframe string) ([]Candle, error)
	SubscribeCandles(security SecurityInfo, timeframe string) error
	LastPrice(security SecurityInfo) (float64, error)
}
