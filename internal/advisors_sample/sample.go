package advisors

import (
	"advisordev/internal/domain"
	"log/slog"
)

func MainAdvisor(name string, stdVolatility float64, logger *slog.Logger) domain.Advisor {
	var advisor = sampleAdvisor()
	return advisor
}

func TestAdvisor(name string) domain.Advisor {
	var advisor = sampleAdvisor()
	return advisor
}

func sampleAdvisor() domain.Advisor {
	return func(c domain.Candle) domain.Advice {
		return domain.Advice{
			SecurityCode: c.SecurityCode,
			DateTime:     c.DateTime,
			Price:        c.ClosePrice,
			Position:     0,
			Details:      "sample",
		}
	}
}
