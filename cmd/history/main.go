package main

import (
	"advisordev/internal/advisors"
	"advisordev/internal/candlestorage"
	"advisordev/internal/core"
	"advisordev/internal/history"
	"fmt"
	"log"
	"os"
	"time"
)

const Concurrency = 6

var logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
var historyCandleService = candlestorage.NewCandleStorage("/Users/vadimchizhov/TradingData/Forts/")

func main() {
	var start = time.Now()
	var err = testSi()
	var elapsed = time.Since(start)
	fmt.Println(elapsed)
	if err != nil {
		fmt.Println(err)
	}
}

func testSi() error {
	const slippage = 0.0002
	var secCodes = QuarterSecurityCodes("Si", 2009, 0, 2022, 3)
	//var secCodes = QuarterSecurityCodes("Si", 2016, 0, 2022, 3)
	var hprs, err = history.MultiContractHprs(historyCandleService, secCodes, func() core.Advisor {
		return advisors.TestAdvisor(logger)
	}, slippage, IsAfterLongHolidays, Concurrency)
	if err != nil {
		return err
	}
	var lever = history.OptimalLever(hprs, history.LimitStDev(0.045))
	hprs = history.HprsWithLever(hprs, lever)
	var stat = history.ComputeHprStatistcs(hprs)
	history.PrintHprReport(stat)
	fmt.Printf("Дох-ть: %.1f Плечо: %.1f\n",
		(stat.MonthHpr-1)*100, lever)
	return nil
}
