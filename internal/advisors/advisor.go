package advisors

import (
	"advisordev/internal/core"
	"log"
)

func TestAdvisor(logger *log.Logger) core.Advisor {
	var advisor = BuyAndHold()
	advisor = ApplyCandleValidation(advisor, logger)
	return advisor
}

func BuyAndHold() core.Advisor {
	return func(candle core.Candle) core.Advice {
		return core.Advice{
			SecurityCode: candle.SecurityCode,
			DateTime:     candle.DateTime,
			Price:        candle.ClosePrice,
			Position:     1.0,
			Details:      "BuyAndHold",
		}
	}
}

//write own advisor here
