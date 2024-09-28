package advisors

import (
	"advisordev/internal/domain"
	"fmt"
	"log/slog"
)

type StrategyConfig struct {
	Name          string  `xml:",attr"`
	SecurityCode  string  `xml:",attr"`
	Lever         float64 `xml:",attr"`
	MaxLever      float64 `xml:",attr"`
	Weight        float64 `xml:",attr"`
	StdVolatility float64 `xml:",attr"`
	Direction     int     `xml:",attr"`
	Position      float64 `xml:",attr"`
}

func Maindvisor(logger *slog.Logger, config StrategyConfig) domain.Advisor {
	var advisor = Sept24ExperimentAdvisor(config.StdVolatility)
	advisor = ApplyCandleValidation(advisor, logger)
	return advisor
}

func TestAdvisor(name string) domain.Advisor {
	const stdVolatility = 0.006
	var advisor domain.Advisor
	if name == "" {
		advisor = Sept24ExperimentAdvisor(stdVolatility)
	} else {
		panic(fmt.Errorf("wrong advisor name %v", name))
	}
	advisor = ApplyCandleValidation(advisor, slog.Default())
	return advisor
}
