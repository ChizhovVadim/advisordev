package main

import (
	"advisordev/internal/candles"
	"advisordev/internal/cli"
	"advisordev/internal/domain"
	"advisordev/internal/history"
	"advisordev/internal/moex"
	"flag"
)

func statusHandler(args []string) error {
	var (
		advisorName   string
		timeframeName string = domain.CandleIntervalMinutes5
		securityName  string
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&advisorName, "advisor", advisorName, "")
	flagset.StringVar(&timeframeName, "timeframe", timeframeName, "")
	flagset.StringVar(&securityName, "security", securityName, "")
	flagset.Parse(args)

	var candleStorage = candles.NewCandleStorage(cli.MapPath("~/TradingData"), timeframeName, moex.TimeZone)
	return history.AdvisorStatus(candleStorage, advisorName, securityName)
}
