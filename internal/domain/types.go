package domain

import "time"

const (
	CandleIntervalMinutes5 = "minutes5"
	CandleIntervalHourly   = "hourly"
	CandleIntervalDaily    = "daily"
)

type Candle struct {
	SecurityCode string
	DateTime     time.Time
	OpenPrice    float64
	HighPrice    float64
	LowPrice     float64
	ClosePrice   float64
	Volume       float64
}

type Advice struct {
	Advisor      string
	SecurityCode string // сюда пишем SecurityCode или SecurityName?
	DateTime     time.Time
	Price        float64
	Position     float64
	Details      any
}

type Advisor func(Candle) Advice

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

type Order struct {
	Portfolio PortfolioInfo
	Security  SecurityInfo
	Volume    int
	Price     float64
}
