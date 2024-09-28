package trader

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
	PricePrecision int
	Lever          float64
}
