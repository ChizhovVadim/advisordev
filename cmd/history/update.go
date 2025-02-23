package main

import (
	"advisordev/internal/candles"
	"advisordev/internal/candles/update"
	"advisordev/internal/cli"
	"advisordev/internal/domain"
	"advisordev/internal/moex"
	"flag"
	"fmt"
	"math"
	"strings"
	"time"
)

func updateHandler(args []string) error {
	var (
		providerName  string
		timeframeName string = domain.CandleIntervalMinutes5
		securityName  string
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&providerName, "provider", providerName, "")
	flagset.StringVar(&timeframeName, "timeframe", timeframeName, "")
	flagset.StringVar(&securityName, "security", securityName, "")
	flagset.Parse(args)

	if securityName == "" {
		return fmt.Errorf("security required")
	}
	var securityCodes = strings.Split(securityName, ",")

	settings, err := loadSettings(cli.MapPath("~/Projects/advisordev/advisor.xml"))
	if err != nil {
		return err
	}

	var candleStorage = candles.NewCandleStorage(cli.MapPath("~/TradingData"), timeframeName, moex.TimeZone)

	var candleProviders []update.ICandleProvider
	candleProvider, err := update.NewCandleProvider(providerName, settings.SecurityCodes, timeframeName, moex.TimeZone)
	if err != nil {
		return err
	}
	candleProviders = append(candleProviders, candleProvider)

	return update.UpdateGroup(securityCodes, candleProviders, candleStorage, calcStartDate, checkPriceChange, 30)
}

func calcStartDate(securityCode string) time.Time {
	// Для квартального фьючерса качаем за 4 месяца до примерной экспирации
	return moex.ApproxExpirationDate(securityCode).AddDate(0, -4, 0)
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
