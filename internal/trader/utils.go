package trader

import (
	"advisordev/internal/domain"
	"advisordev/internal/quik"
	"advisordev/internal/utils"
	"fmt"
	"strings"
	"time"
)

const FuturesClassCode = "SPBFUT"
const CurrencyClassCode = "CETS"

func isToday(d time.Time) bool {
	var y1, m1, d1 = d.Date()
	var y2, m2, d2 = time.Now().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func getSecurityInfoHardCode(securityName string) (SecurityInfo, error) {
	if securityName == "CNYRUB_TOM" {
		return SecurityInfo{
			Name:      securityName,
			ClassCode: CurrencyClassCode,
			Code:      securityName,
		}, nil
	}
	if securityName == "CNYRUBF" {
		return SecurityInfo{
			Name:           securityName,
			ClassCode:      FuturesClassCode,
			Code:           securityName,
			PricePrecision: 3,
			PriceStep:      0.001,
			PriceStepCost:  1,
			Lever:          1000,
		}, nil
	}
	if strings.HasPrefix(securityName, "CNY") {
		securityCode, err := utils.EncodeSecurity(securityName)
		if err != nil {
			return SecurityInfo{}, err
		}
		return SecurityInfo{
			Name:           securityName,
			ClassCode:      FuturesClassCode,
			Code:           securityCode,
			PricePrecision: 3,
			PriceStep:      0.001,
			PriceStepCost:  1,
			Lever:          1000,
		}, nil
	}
	return SecurityInfo{}, fmt.Errorf("secInfo not found %v", securityName)
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
