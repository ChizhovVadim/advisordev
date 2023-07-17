package advisors

import (
	"advisordev/internal/core"
	"math"
)

type IUpdateIndicator interface {
	Add(candle core.Candle)
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

type MainSessionRebalanceInd struct {
	value bool
}

func (ind *MainSessionRebalanceInd) Add(candle core.Candle) {
	ind.value = core.IsMainFortsSession(candle.DateTime)
}

func (ind *MainSessionRebalanceInd) Value() bool {
	return ind.value
}

type VolatilityInd struct {
	standardVol float64
	period      int
	buffer      []float64
	lastCandle  core.Candle
}

func (ind *VolatilityInd) Add(candle core.Candle) {
	if isMainSessionOneDayCandles(ind.lastCandle, candle) {
		var change = math.Log(candle.ClosePrice / ind.lastCandle.ClosePrice)
		ind.buffer = append(ind.buffer, change)
	}
	ind.lastCandle = candle
}

func (ind *VolatilityInd) Value() float64 {
	if len(ind.buffer) < ind.period/2 {
		return ind.standardVol
	}
	var skip = len(ind.buffer) - ind.period
	if skip > 0 {
		ind.buffer = ind.buffer[skip:]
	}
	return math.Sqrt(float64(len(ind.buffer))) * core.StDev(ind.buffer)
}

// Учебный пример
type MovingAverageIng struct {
	period int
	prices []float64
	value  float64
}

func (ind *MovingAverageIng) Add(candle core.Candle) {
	if core.IsMainFortsSession(candle.DateTime) {
		ind.prices = append(ind.prices, candle.ClosePrice)
		if len(ind.prices) > ind.period {
			ind.prices = ind.prices[len(ind.prices)-ind.period:]
		}
		ind.value = compare(mean(ind.prices), candle.ClosePrice)
	}
}

func (ind *MovingAverageIng) Value() float64 {
	return ind.value
}
