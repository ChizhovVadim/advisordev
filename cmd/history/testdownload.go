package main

import (
	"advisordev/internal/candles/update"
	"advisordev/internal/cli"
	"advisordev/internal/domain"
	"advisordev/internal/moex"
	"flag"
	"fmt"
	"log/slog"
	"time"
)

func testDownloadHandler(args []string) error {
	var (
		providerName  string
		timeframeName string = domain.CandleIntervalMinutes5
		securityName  string
		startDate     cli.DateValue = cli.DateValue{Date: time.Now().AddDate(0, 0, -5)}
		finishDate    cli.DateValue = cli.DateValue{Date: time.Now()}
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&providerName, "provider", providerName, "")
	flagset.StringVar(&timeframeName, "timeframe", timeframeName, "")
	flagset.StringVar(&securityName, "security", securityName, "")
	flagset.Var(&startDate, "start", "")
	flagset.Var(&finishDate, "finish", "")
	flagset.Parse(args)

	if securityName == "" {
		return fmt.Errorf("security required")
	}

	settings, err := loadSettings(cli.MapPath("~/Projects/advisordev/advisor.xml"))
	if err != nil {
		return err
	}

	provider, err := update.NewCandleProvider(
		providerName, settings.SecurityCodes, timeframeName, moex.TimeZone)
	if err != nil {
		return err
	}
	candles, err := provider.Load(securityName, startDate.Date, finishDate.Date)
	if err != nil {
		return err
	}
	slog.Info("Downloaded",
		"size", len(candles),
		"head", candles[:min(5, len(candles))],
		"tail", candles[max(0, len(candles)-5):],
	)
	return nil
}
