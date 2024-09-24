package main

import (
	"advisordev/internal/advisors"
	"advisordev/internal/candles"
	"advisordev/internal/domain"
	"advisordev/internal/history"
	"advisordev/internal/utils"
	"flag"
)

func statusHandler(args []string) error {
	var (
		advisorName   = ""
		timeframeName = candles.TFMinutes5
		securityName  = "CNY-12.24"
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&advisorName, "advisor", advisorName, "")
	flagset.StringVar(&timeframeName, "timeframe", timeframeName, "")
	flagset.StringVar(&securityName, "security", securityName, "")
	flagset.Parse(args)

	var advisorBuilder = func() domain.Advisor {
		return advisors.TestAdvisor(advisorName)
	}
	var historyCandleStorage = candles.NewCandleStorage(utils.MapPath("~/TradingData"), timeframeName, utils.Moscow)
	return history.AdvisorStatus(historyCandleStorage, securityName, advisorBuilder)
}
