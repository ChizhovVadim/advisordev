package advisors

import "advisordev/internal/domain"

type IUpdateIndicator interface {
	Add(candle domain.Candle)
}

type IBoolIndicator interface {
	Value() bool
}

type IFloat64Indicator interface {
	Value() float64
}

type IUpdateBoolIndicator interface {
	IUpdateIndicator
	IBoolIndicator
}

type IUpdateFloat64Indicator interface {
	IUpdateIndicator
	IFloat64Indicator
}
