package quik

import (
	"advisordev/internal/domain"
	"advisordev/internal/moex"
	"fmt"
	"math"
	"strconv"
	"time"
)

func formatPrice(priceStep float64, pricePrecision int, price float64) string {
	if priceStep != 0 {
		price = math.Round(price/priceStep) * priceStep
	}
	return strconv.FormatFloat(price, 'f', pricePrecision, 64)
}

func quikTimeframe(timeframe string) (CandleInterval, error) {
	if timeframe == domain.CandleIntervalMinutes5 {
		return CandleIntervalM5, nil
	}
	return 0, fmt.Errorf("timeframe not supported %v", timeframe)
}

func convertQuikCandle(candle Candle) domain.Candle {
	return domain.Candle{
		SecurityCode: candle.SecCode,
		DateTime:     convertQuikDateTime(candle.Datetime, moex.TimeZone),
		OpenPrice:    candle.Open,
		HighPrice:    candle.High,
		LowPrice:     candle.Low,
		ClosePrice:   candle.Close,
		Volume:       candle.Volume,
	}
}

func convertQuikDateTime(t QuikDateTime, loc *time.Location) time.Time {
	//TODO ms
	return time.Date(t.Year, time.Month(t.Month), t.Day, t.Hour, t.Min, t.Sec, 0, loc)
}

func isToday(d time.Time) bool {
	var y1, m1, d1 = d.Date()
	var y2, m2, d2 = time.Now().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func AsInt(a any) (int, error) {
	switch v := a.(type) {
	case float64:
		return int(v), nil
	case string:
		var f, err = strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("wrong value type %v", v)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("wrong value type %v", v)
	}
}

func AsFloat64(a any) (float64, error) {
	switch v := a.(type) {
	case float64:
		return v, nil
	case string:
		var f, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("wrong value type %v", v)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("wrong value type %v", v)
	}
}
