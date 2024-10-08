package trader

import "advisordev/internal/quik"

type IStrategy interface {
	OnNewCandle(newCandle quik.Candle) (bool, error)
	CheckPosition() error
}

type Client struct {
	Key            string  `xml:",attr"`
	Firm           string  `xml:",attr"`
	Portfolio      string  `xml:",attr"`
	PublishCandles bool    `xml:",attr"`
	Amount         float64 `xml:",attr"`
	MaxAmount      float64 `xml:",attr"`
	Weight         float64 `xml:",attr"`
	Port           int     `xml:",attr"`
}

type SecurityInfo struct {
	// точность (кол-во знаков после запятой). Если шаг цены может быть не круглым (0.05), то этого будет недостаточно.
	PricePrecision int
	// шаг цены
	PriceStep float64
	// число базового актива в контракте
	Lever float64
}
