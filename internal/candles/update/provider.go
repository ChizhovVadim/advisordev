package update

import (
	"advisordev/internal/domain"
	"fmt"
	"net/http"
	"time"
)

type SecurityCode struct {
	Code      string `xml:",attr"`
	FinamCode string `xml:",attr"`
	MfdCode   string `xml:",attr"`
}

type ICandleProvider interface {
	Name() string
	Load(securityName string, beginDate, endDate time.Time) ([]domain.Candle, error)
}

func NewCandleProvider(
	key string,
	secCodes []SecurityCode,
	candleInterval string,
	loc *time.Location,
) (ICandleProvider, error) {
	if key == "finam" {
		return NewFinam(
			prepareCodes(secCodes, func(sc SecurityCode) string { return sc.FinamCode }),
			candleInterval,
			&http.Client{Timeout: 25 * time.Second},
			loc)
	}
	if key == "mfd" {
		return NewMfd(
			prepareCodes(secCodes, func(sc SecurityCode) string { return sc.MfdCode }),
			candleInterval,
			&http.Client{Timeout: 25 * time.Second},
			loc)
	}
	return nil, fmt.Errorf("bad provider %v", key)
}

func prepareCodes(
	secCodes []SecurityCode,
	key func(SecurityCode) string,
) map[string]string {
	var result = make(map[string]string)
	for _, code := range secCodes {
		providerCode := key(code)
		if providerCode != "" {
			result[code.Code] = providerCode
		}
	}
	return result
}
