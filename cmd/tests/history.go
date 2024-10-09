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

type HistoryTest struct {
	advisorName   string
	timeframeName string
	securityName  string
	lever         float64
	slippage      float64
	startYear     int
	startQuarter  int
	finishYear    int
	finishQuarter int
	multiContract bool
}

func historyHandler(args []string) error {
	var today = time.Now()
	var t = HistoryTest{
		advisorName:   "",
		timeframeName: candles.TFMinutes5,
		securityName:  "CNY",
		lever:         9.0,
		slippage:      0.0002,
		startYear:     today.Year(),
		startQuarter:  0,
		finishYear:    today.Year(),
		finishQuarter: 3,
		multiContract: true,
	}
	var concurrency = runtime.NumCPU()

	var flagset = flag.NewFlagSet("", flag.ExitOnError)
	flagset.StringVar(&t.advisorName, "advisor", t.advisorName, "")
	flagset.StringVar(&t.timeframeName, "timeframe", t.timeframeName, "")
	flagset.StringVar(&t.securityName, "security", t.securityName, "")
	flagset.Float64Var(&t.lever, "lever", t.lever, "")
	flagset.Float64Var(&t.slippage, "slippage", t.slippage, "")
	flagset.IntVar(&t.startYear, "startyear", t.startYear, "")
	flagset.IntVar(&t.startQuarter, "startquarter", t.startQuarter, "")
	flagset.BoolVar(&t.multiContract, "multy", t.multiContract, "")
	flagset.Parse(args)

	var start = time.Now()
	defer func() {
		fmt.Println("Elapsed:", time.Since(start))
	}()

	hprs, err := calcDailyHprs(t, concurrency)
	if err != nil {
		return err
	}

	var stat = history.ComputeHprStatistcs(hprs)
	history.PrintHprReport(stat)
	return nil
}

func calcDailyHprs(t HistoryTest, concurrency int) ([]history.DateSum, error) {
	var advisorBuilder = func() domain.Advisor {
		return advisors.TestAdvisor(t.advisorName)
	}
	var historyCandleStorage = candles.NewCandleStorage(utils.MapPath("~/TradingData"), t.timeframeName, utils.Moscow)

	var hprs []history.DateSum
	if t.multiContract {
		var tr = utils.TimeRange{
			StartYear:     t.startYear,
			StartQuarter:  t.startQuarter,
			FinishYear:    t.finishYear,
			FinishQuarter: t.finishQuarter,
		}
		var tickers = utils.QuarterSecurityCodes(t.securityName, tr)
		var err error
		hprs, err = history.MultiContractHprs(historyCandleStorage, tickers, advisorBuilder, t.slippage, history.IsAfterLongHolidays, concurrency)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		hprs, err = history.SingleContractHprs(historyCandleStorage, t.securityName, advisorBuilder, t.slippage, history.IsAfterLongHolidays)
		if err != nil {
			return nil, err
		}
	}
	var lever = t.lever
	if lever == 0 {
		lever = history.OptimalLever(hprs, history.LimitStDev(0.045))
		fmt.Printf("Плечо: %.1f\n", lever)
	}
	hprs = history.HprsWithLever(hprs, lever)
	return hprs, nil
}
