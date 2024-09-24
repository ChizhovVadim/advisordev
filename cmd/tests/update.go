package main

import (
	"advisordev/internal/candles"
	"advisordev/internal/domain"
	"advisordev/internal/utils"
	"flag"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

func updateHandler(args []string) error {
	var (
		providerName  string
		timeframeName string = candles.TFMinutes5
		securityName  string
		startDate     utils.DateValue
		maxDays       int
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&providerName, "provider", providerName, "")
	flagset.StringVar(&timeframeName, "timeframe", timeframeName, "")
	flagset.StringVar(&securityName, "security", securityName, "")
	flagset.Var(&startDate, "start", "")
	flagset.IntVar(&maxDays, "maxdays", maxDays, "")
	flagset.Parse(args)

	if securityName == "" {
		return fmt.Errorf("security required")
	}
	var strategySettings, err = loadStrategySettings(utils.MapPath("~/Projects/advisordev/advisor.xml"))
	if err != nil {
		return err
	}
	canldeProviders, err := buildCandleProviders(providerName, timeframeName, strategySettings.SecurityCodes)
	if err != nil {
		return err
	}
	var candleStorage = candles.NewCandleStorage(utils.MapPath("~/TradingData"), timeframeName, utils.Moscow)
	var startDateFunc func(string) time.Time
	if !startDate.Date.IsZero() {
		startDateFunc = func(string) time.Time {
			return startDate.Date
		}
	} else {
		startDateFunc = calcStartDate
	}
	var securityCodes = strings.Split(securityName, ",")
	return candles.UpdateGroup(securityCodes, canldeProviders, candleStorage, startDateFunc, checkPriceChange, maxDays)
}

func calcStartDate(securityCode string) time.Time {
	return utils.ApproxExpirationDate(securityCode).AddDate(0, -1, 0)
}

func checkPriceChange(x, y domain.Candle) error {
	const Width = 0.25
	var closeChange = math.Abs(math.Log(x.ClosePrice / y.ClosePrice))
	var openChange = math.Abs(math.Log(x.ClosePrice / y.OpenPrice))
	if openChange >= Width && closeChange >= Width {
		return fmt.Errorf("big jump %v %v", x, y)
	}
	return nil
}

func buildCandleProviders(
	providerName string,
	timeframeName string,
	securityCodes []SecurityCode,
) ([]candles.ICandleProvider, error) {
	//TODO maybe strings.Split(providerName)?
	var finamCodes = make(map[string]string)
	var mfdCodes = make(map[string]string)
	for _, secCode := range securityCodes {
		if secCode.FinamCode != "" {
			finamCodes[secCode.Code] = secCode.FinamCode
		}
		if secCode.MfdCode != "" {
			mfdCodes[secCode.Code] = secCode.MfdCode
		}
	}
	var client = &http.Client{
		Timeout: 25 * time.Second,
	}
	var result []candles.ICandleProvider
	if providerName == "" || providerName == "finam" {
		var provider, err = candles.NewFinam(finamCodes, timeframeName, client)
		if err != nil {
			return nil, err
		}
		result = append(result, provider)
	}
	if providerName == "" || providerName == "mfd" {
		var provider, err = candles.NewMfd(mfdCodes, timeframeName, client)
		if err != nil {
			return nil, err
		}
		result = append(result, provider)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("providers empty")
	}
	return result, nil
}
