package trader

import (
	"encoding/xml"
	"os"
)

type TraderConfig struct {
	Clients []ClientConfig `xml:"Client"`
	Signals []SignalConfig `xml:"Signal"`
}

type ClientConfig struct {
	Key        string            `xml:",attr"`
	Type       string            `xml:",attr"`
	Port       int               `xml:",attr"`
	MarketData bool              `xml:",attr"`
	Portfolios []PortfolioConfig `xml:"Portfolio"`
}

type PortfolioConfig struct {
	Firm       string           `xml:",attr"`
	Portfolio  string           `xml:"Account,attr"`
	MaxAmount  float64          `xml:",attr"`
	Strategies []StrategyConfig `xml:"Strategy"`
}

type StrategyConfig struct {
	Advisor  string  `xml:",attr"`
	Security string  `xml:",attr"`
	Weight   float64 `xml:",attr"`
}

type SignalConfig struct {
	Advisor       string  `xml:",attr"`
	Security      string  `xml:",attr"`
	Lever         float64 `xml:",attr"`
	MaxLever      float64 `xml:",attr"`
	StdVolatility float64 `xml:",attr"`
	Weight        float64 `xml:",attr"`
}

func LoadConfig(filePath string) (TraderConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return TraderConfig{}, err
	}
	defer file.Close()

	var result TraderConfig
	err = xml.NewDecoder(file).Decode(&result)
	if err != nil {
		return TraderConfig{}, err
	}
	return result, nil
}
