package main

import (
	"advisordev/internal/candles"
	"advisordev/internal/history"
	"runtime"
)

func portfolioHandler([]string) error {
	var concurrency = runtime.NumCPU()

	hprsCNY, err := calcDailyHprs(HistoryTest{
		securityName:  "CNY",
		advisorName:   "",
		timeframeName: candles.TFMinutes5,
		multiContract: true,
		startYear:     2024,
		startQuarter:  0,
		finishYear:    2024,
		finishQuarter: 3,
		lever:         9.0,
		slippage:      0.0002,
	}, concurrency)
	if err != nil {
		return nil
	}

	hprsGold, err := calcDailyHprs(HistoryTest{
		securityName:  "GLDRUB_TOM",
		advisorName:   "",
		timeframeName: candles.TFMinutes5,
		multiContract: false,
		lever:         9.0,
		slippage:      0.0002,
	}, concurrency)
	if err != nil {
		return nil
	}

	var portfolio = [][]history.DateSum{hprsCNY, hprsGold}
	var weights = []float64{0.5, 0.5}

	var totalHprs = history.CombineHprs(portfolio, weights)
	history.PrintHprReport(history.ComputeHprStatistcs(totalHprs))
	return nil
}
