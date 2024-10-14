package trader

import (
	"advisordev/internal/domain"
	"advisordev/internal/utils"
	"fmt"
	"strings"
)

const (
	FuturesClassCode  = domain.FuturesClassCode
	CurrencyClassCode = "CETS"
	// Акции и ДР
	StockClassCode = "TQBR"
	// ETF
	ETFClassCode = "TQTF"
)

func getSecurityInfoHardCode(securityName string) (domain.SecurityInfo, error) {
	if securityName == "CNYRUB_TOM" {
		return domain.SecurityInfo{
			Name:      securityName,
			ClassCode: CurrencyClassCode,
			Code:      securityName,
		}, nil
	}
	if securityName == "CNYRUBF" {
		return domain.SecurityInfo{
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
			return domain.SecurityInfo{}, err
		}
		return domain.SecurityInfo{
			Name:           securityName,
			ClassCode:      FuturesClassCode,
			Code:           securityCode,
			PricePrecision: 3,
			PriceStep:      0.001,
			PriceStepCost:  1,
			Lever:          1000,
		}, nil
	}
	return domain.SecurityInfo{}, fmt.Errorf("secInfo not found %v", securityName)
}

func priceWithSlippage(price float64, volume int) float64 {
	const Slippage = 0.001
	if volume > 0 {
		return price * (1 + Slippage)
	} else {
		return price * (1 - Slippage)
	}
}
