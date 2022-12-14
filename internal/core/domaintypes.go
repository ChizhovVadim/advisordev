package core

import (
	"errors"
	"time"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNoData         = errors.New("no data")
)

type Candle struct {
	SecurityCode string
	DateTime     time.Time
	OpenPrice    float64
	HighPrice    float64
	LowPrice     float64
	ClosePrice   float64
	Volume       float64
}

type Advice struct {
	SecurityCode string
	DateTime     time.Time
	Price        float64
	Position     float64
	Details      interface{}
}

type Advisor func(Candle) Advice
