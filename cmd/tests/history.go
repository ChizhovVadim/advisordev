package main

import (
	"advisordev/internal/advisors"
	"advisordev/internal/candles"
	"advisordev/internal/domain"
	"advisordev/internal/history"
	"advisordev/internal/utils"
	"flag"
	"fmt"
	"runtime"
	"time"
)

func historyHandler(args []string) error {
	var (
		advisorName   = ""
		timeframeName = candles.TFMinutes5
		securityName  = "CNY"
		lever         = 9.0
		slippage      = 0.0002
		startYear     = time.Now().Year()
		startQuarter  = 0
		finishYear    = startYear
		finishQuarter = 3
		multiContract = true
		concurrency   = runtime.NumCPU()
	)

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&advisorName, "advisor", advisorName, "")
	flagset.StringVar(&timeframeName, "timeframe", timeframeName, "")
	flagset.StringVar(&securityName, "security", securityName, "")
	flagset.Float64Var(&lever, "lever", lever, "")
	flagset.Float64Var(&slippage, "slippage", slippage, "")
	flagset.IntVar(&startYear, "startyear", startYear, "")
	flagset.IntVar(&startQuarter, "startquarter", startQuarter, "")
	flagset.BoolVar(&multiContract, "multy", multiContract, "")
	flagset.Parse(args)

	var start = time.Now()
	defer func() {
		fmt.Println("Elapsed:", time.Since(start))
	}()

	var advisorBuilder = func() domain.Advisor {
		return advisors.TestAdvisor(advisorName)
	}
	var historyCandleStorage = candles.NewCandleStorage(utils.MapPath("~/TradingData"), timeframeName, utils.Moscow)

	var hprs []history.DateSum
	if multiContract {
		var tr = utils.TimeRange{
			StartYear:     startYear,
			StartQuarter:  startQuarter,
			FinishYear:    finishYear,
			FinishQuarter: finishQuarter,
		}
		var tickers = utils.QuarterSecurityCodes(securityName, tr)
		var err error
		hprs, err = history.MultiContractHprs(historyCandleStorage, tickers, advisorBuilder, slippage, history.IsAfterLongHolidays, concurrency)
		if err != nil {
			return err
		}
	} else {
		var err error
		hprs, err = history.SingleContractHprs(historyCandleStorage, securityName, advisorBuilder, slippage, history.IsAfterLongHolidays)
		if err != nil {
			return err
		}
	}
	if lever == 0 {
		lever = history.OptimalLever(hprs, history.LimitStDev(0.045))
		fmt.Printf("Плечо: %.1f\n", lever)
	}
	hprs = history.HprsWithLever(hprs, lever)
	var stat = history.ComputeHprStatistcs(hprs)
	history.PrintHprReport(stat)
	return nil
}
