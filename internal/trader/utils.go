package trader

import (
	"advisordev/internal/domain"
	"advisordev/internal/quik"
	"advisordev/internal/utils"
	"strings"
	"time"
)

const ClassCode = "SPBFUT"

func isToday(d time.Time) bool {
	var y1, m1, d1 = d.Date()
	var y2, m2, d2 = time.Now().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func getSecurityInfo(security string) SecurityInfo {
	//TODO hack
	if strings.HasPrefix(security, "CNY") {
		return SecurityInfo{
			PricePrecision: 3,
			Lever:          1000,
		}
	}
	return SecurityInfo{
		PricePrecision: 0,
		Lever:          1,
	}
}

func convertQuikCandle(security string, candle quik.Candle) domain.Candle {
	return domain.Candle{
		SecurityCode: security,
		DateTime:     convertQuikDateTime(candle.Datetime, utils.Moscow),
		OpenPrice:    candle.Open,
		HighPrice:    candle.High,
		LowPrice:     candle.Low,
		ClosePrice:   candle.Close,
		Volume:       candle.Volume,
	}
}

func convertQuikDateTime(t quik.QuikDateTime, loc *time.Location) time.Time {
	//TODO ms
	return time.Date(t.Year, time.Month(t.Month), t.Day, t.Hour, t.Min, t.Sec, 0, loc)
}
