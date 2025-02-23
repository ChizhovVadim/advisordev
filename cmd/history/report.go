package main

import (
	"advisordev/internal/candles"
	"advisordev/internal/cli"
	"advisordev/internal/domain"
	"advisordev/internal/history"
	"advisordev/internal/moex"
	"flag"
	"time"
)

func reportHandler(args []string) error {
	var today = time.Now()
	var (
		advisorName   string
		timeframeName string = domain.CandleIntervalMinutes5
		securityName  string
		lever         float64
		slippage      float64 = 0.0002
		startYear     int     = today.Year()
		startQuarter  int     = 0
		finishYear    int     = today.Year()
		finishQuarter int     = 3
		multiContract bool    = true
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&advisorName, "advisor", advisorName, "")
	flagset.StringVar(&timeframeName, "timeframe", timeframeName, "")
	flagset.StringVar(&securityName, "security", securityName, "")
	flagset.Float64Var(&lever, "lever", lever, "")
	flagset.Float64Var(&slippage, "slippage", slippage, "")
	flagset.IntVar(&startYear, "startyear", startYear, "")
	flagset.IntVar(&startQuarter, "startquarter", startQuarter, "")
	flagset.IntVar(&finishYear, "finishyear", finishYear, "")
	flagset.IntVar(&finishQuarter, "finishquarter", finishQuarter, "")
	flagset.BoolVar(&multiContract, "multy", multiContract, "")
	flagset.Parse(args)

	var candleStorage = candles.NewCandleStorage(cli.MapPath("~/TradingData"), timeframeName, moex.TimeZone)
	return history.AdvisorReport(candleStorage, advisorName, securityName, lever, slippage, startYear, startQuarter, finishYear, finishQuarter, multiContract)
}
