package advisors

import (
	"advisordev/internal/domain"
	"fmt"
	"log/slog"
)

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
