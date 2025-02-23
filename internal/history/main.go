package history

import (
	advisors "advisordev/internal/advisors_sample"
	"advisordev/internal/domain"
	"advisordev/internal/moex"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func AdvisorStatus(
	candleStorage domain.ICandleStorage,
	advisorName string,
	securityName string,
) error {
	var advisor = advisors.TestAdvisor(advisorName)
	var advices []domain.Advice
	for candle, err := range candleStorage.Candles(securityName) {
		if err != nil {
			return err
		}
		var advice = advisor(candle)
		if advice.DateTime.IsZero() {
			continue
		}
		if len(advices) == 0 ||
			advices[len(advices)-1].Position != advice.Position {
			advices = append(advices, advice)
		}
	}
	const N = 10
	for _, item := range advices[max(0, len(advices)-N):] {
		fmt.Println(item)
	}
	return nil
}

func AdvisorReport(
	candleStorage domain.ICandleStorage,
	advisorName string,
	securityName string,
	lever float64,
	slippage float64,
	startYear int,
	startQuarter int,
	finishYear int,
	finishQuarter int,
	multiContract bool,
) error {
	var start = time.Now()
	defer func() {
		fmt.Println("Elapsed:", time.Since(start))
	}()

	var secCodes []string
	if multiContract {
		var tr = moex.TimeRange{
			StartYear:     startYear,
			StartQuarter:  startQuarter,
			FinishYear:    finishYear,
			FinishQuarter: finishQuarter,
		}
		secCodes = moex.QuarterSecurityCodes(securityName, tr)
	} else {
		secCodes = []string{securityName}
	}

	var hprs, err = MultiContractHprs(
		candleStorage, advisorName, secCodes, slippage, isAfterLongHolidays, runtime.NumCPU())
	if err != nil {
		return err
	}
	if lever == 0 {
		lever = optimalLever(hprs, limitStDev(0.045))
	}
	hprs = hprsWithLever(hprs, lever)

	fmt.Println("Отчет", advisorName, securityName)
	fmt.Printf("Плечо: %.1f\n", lever)
	ReportDailyResults(hprs)
	return nil
}

func limitStDev(stDev float64) func([]DateSum) bool {
	return func(source []DateSum) bool {
		return stDevHprs(source) <= stDev
	}
}

func optimalLever(hprs []DateSum, riskSpecification func([]DateSum) bool) float64 {
	var minHpr = hprs[0].Sum
	for _, x := range hprs[1:] {
		if x.Sum < minHpr {
			minHpr = x.Sum
		}
	}
	var maxLever = 1.0 / (1.0 - minHpr)
	var bestHpr = 1.0
	var bestLever = 0.0
	const step = 0.001

	for ratio := step; ratio <= 1; ratio += step {
		var lever = maxLever * ratio
		var leverHprs = hprsWithLever(hprs, lever)
		if !riskSpecification(leverHprs) {
			break
		}
		var hpr = totalHpr(leverHprs)
		if hpr < bestHpr {
			break
		}
		bestHpr = hpr
		bestLever = lever
	}

	return bestLever
}

func hprsWithLever(source []DateSum, lever float64) []DateSum {
	var result = make([]DateSum, len(source))
	for i, item := range source {
		result[i] = DateSum{
			Date: item.Date,
			Sum:  1 + lever*(item.Sum-1),
		}
	}
	return result
}

func MultiContractHprs(
	candleStorage domain.ICandleStorage,
	advisorName string,
	secCodes []string,
	slippage float64,
	skipPnl func(time.Time, time.Time) bool,
	concurrency int,
) ([]DateSum, error) {
	if len(secCodes) == 1 {
		return SingleContractHprs(
			candleStorage.Candles(secCodes[0]),
			advisors.TestAdvisor(advisorName),
			slippage,
			skipPnl)
	}

	var index int32 = -1
	var wg = &sync.WaitGroup{}
	var hprsByContracts = make([][]DateSum, len(secCodes))
	for threadIndex := 0; threadIndex < concurrency; threadIndex++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				var i = int(atomic.AddInt32(&index, 1))
				if i >= len(secCodes) {
					break
				}
				var securityCode = secCodes[i]
				var hprs, err = SingleContractHprs(
					candleStorage.Candles(securityCode),
					advisors.TestAdvisor(advisorName),
					slippage,
					skipPnl)
				if err != nil {
					log.Println(err)
					continue
				}
				hprsByContracts[i] = hprs
			}
		}()
	}
	wg.Wait()

	return concatHprs(hprsByContracts), nil
}

func concatHprs(hprsByContracts [][]DateSum) []DateSum {
	var result []DateSum
	for _, hprs := range hprsByContracts {
		if len(hprs) == 0 {
			continue
		}

		if len(result) != 0 {
			// последний день предыдущего контракта может быть не полный
			result = result[:len(result)-1]
		}

		var last = time.Time{}
		if len(result) != 0 {
			last = result[len(result)-1].Date
		}
		for i := 0; i < len(hprs); i++ {
			if hprs[i].Date.After(last) {
				result = append(result, hprs[i:]...)
				break
			}
		}
	}
	return result
}
