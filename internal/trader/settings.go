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

type StrategyConfig struct {
	Trader        string  `xml:",attr"`
	Name          string  `xml:",attr"` //TODO rename->Advisor
	SecurityCode  string  `xml:",attr"`
	Lever         float64 `xml:",attr"`
	MaxLever      float64 `xml:",attr"`
	Weight        float64 `xml:",attr"`
	StdVolatility float64 `xml:",attr"`
	Direction     int     `xml:",attr"`
	//Position      float64 `xml:",attr"`
}
