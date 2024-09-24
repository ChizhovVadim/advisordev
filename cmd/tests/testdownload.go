package main

import (
	"advisordev/internal/candles"
	"advisordev/internal/utils"
	"flag"
	"fmt"
	"log"
	"time"
)

func testDownloadHandler(args []string) error {
	var (
		providerName  string
		timeframeName string = candles.TFMinutes5
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
	var strategySettings, err = loadStrategySettings(utils.MapPath("~/Projects/advisordev/advisor.xml"))
	if err != nil {
		return err
	}
	canldeProviders, err := buildCandleProviders(providerName, timeframeName, strategySettings.SecurityCodes)
	if err != nil {
		return err
	}

	var today = time.Now()
	for _, candleProvider := range canldeProviders {
		var candles, err = candleProvider.Load(securityName, today.AddDate(0, 0, -5), today)
		if err != nil {
			log.Println("Download failed",
				"provider", candleProvider.Name(),
				"securityCode", securityName,
				"err", err)
		} else {
			log.Println("Downloaded",
				"provider", candleProvider.Name(),
				"securityCode", securityName,
				"size", len(candles),
				"head", candles[:utils.Min(5, len(candles))],
				"tail", candles[utils.Max(0, len(candles)-5):],
			)
		}
	}

	return nil
}
