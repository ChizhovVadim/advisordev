package trader

import (
	"advisordev/internal/domain"
)

type IStrategy interface {
	OnNewCandle(newCandle domain.Candle) (bool, error)
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
