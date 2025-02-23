package trader

import (
	"advisordev/internal/domain"
	"math"
)

func applyStdDecorator(
	advisor domain.Advisor,
	direction int,
	lever, maxLever float64,
) domain.Advisor {

	return func(candle domain.Candle) domain.Advice {
		var advice = advisor(candle)
		if advice.DateTime.IsZero() {
			return advice
		}
		var position = advice.Position
		if direction == 1 {
			position = math.Max(position, 0)
		} else if direction == -1 {
			position = math.Min(position, 0)
		}
		position = math.Max(-maxLever, math.Min(maxLever, position*lever))
		return withPosition(advice, position, "StdDecorator")
	}
}

func withPosition(adv domain.Advice, position float64, name string) domain.Advice {
	type AdviceDetails struct {
		Name          string
		ChildPosition float64
		ChildDetails  interface{}
	}

	const LogDetails = true
	if adv.Position == position {
		return adv
	}
	var result = domain.Advice{
		SecurityCode: adv.SecurityCode,
		DateTime:     adv.DateTime,
		Price:        adv.Price,
		Position:     position}
	if LogDetails {
		result.Details = AdviceDetails{
			Name:          name,
			ChildPosition: adv.Position,
			ChildDetails:  adv.Details,
		}
	}
	return result
}
