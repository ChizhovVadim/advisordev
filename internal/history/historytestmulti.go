package history

import (
	"advisordev/internal/domain"
	"advisordev/internal/utils"
	"iter"
	"log"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func HprsWithLever(source []DateSum, lever float64) []DateSum {
	var result = make([]DateSum, len(source))
	for i, item := range source {
		result[i] = DateSum{
			Date: item.Date,
			Sum:  1 + lever*(item.Sum-1),
		}
	}
	return result
}

func LimitStDev(stDev float64) func([]DateSum) bool {
	return func(source []DateSum) bool {
		return stDevHprs(source) <= stDev
	}
}

func OptimalLever(hprs []DateSum, riskSpecification func([]DateSum) bool) float64 {
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
		var leverHprs = HprsWithLever(hprs, lever)
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

func CombineHprs(hprs [][]DateSum, weights []float64) []DateSum {
	var start = hprs[0][0].Date
	var finish = hprs[0][len(hprs[0])-1].Date
	for i := 1; i < len(hprs); i++ {
		if d := hprs[i][0].Date; d.After(start) {
			start = d
		}
		if d := hprs[i][len(hprs[i])-1].Date; d.Before(finish) {
			finish = d
		}
	}
	var m = make(map[time.Time]float64)
	for i := range hprs {
		var w = weights[i]
		for _, ds := range hprs[i] {
			if !ds.Date.Before(start) && !ds.Date.After(finish) {
				m[ds.Date] += w * (ds.Sum - 1)
			}
		}
	}
	var result = make([]DateSum, 0, len(m))
	for k, v := range m {
		result = append(result, DateSum{k, 1 + v})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date.Before(result[j].Date)
	})
	return result
}

func ComputeCorrelation(hprsA, hprsB []DateSum) float64 {
	var data []point2D
	var i, j int
	for i < len(hprsA) && j < len(hprsB) {
		var hprA, hprB = hprsA[i], hprsB[j]
		if hprA.Date == hprB.Date {
			data = append(data, point2D{hprA.Sum, hprB.Sum})
			i++
			j++
		} else if hprA.Date.Before(hprB.Date) {
			i++
		} else {
			j++
		}
	}
	return pearson(data)
}

type point2D struct {
	X, Y float64
}

func pearson(source []point2D) float64 {
	var sumA = 0.0
	var sumB = 0.0
	for i := range source {
		sumA += source[i].X
		sumB += source[i].Y
	}
	var meanA = sumA / float64(len(source))
	var meanB = sumB / float64(len(source))
	var sumAB = 0.0
	sumA = 0.0
	sumB = 0.0
	for i := range source {
		sumAB += (source[i].X - meanA) * (source[i].Y - meanB)
		sumA += (source[i].X - meanA) * (source[i].X - meanA)
		sumB += (source[i].Y - meanB) * (source[i].Y - meanB)
	}
	return sumAB / math.Sqrt(sumA*sumB)
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

type ICandleStorage interface {
	Walk(securityCode string, onCandle func(domain.Candle) error) error
	Candles(securityCode string) iter.Seq2[domain.Candle, error]
}

func SingleContractHprs(
	historyCandleService ICandleStorage,
	securityCode string,
	advisorBuilder func() domain.Advisor,
	slippage float64,
	skipPnl func(time.Time, time.Time) bool) ([]DateSum, error) {

	var result []DateSum
	var pnl = 0.0
	var baseAdvice = domain.Advice{}
	var lastAdvice = domain.Advice{}
	var advisor = advisorBuilder()

	for candle, err := range historyCandleService.Candles(securityCode) {
		if err != nil {
			return nil, err
		}
		var advice = advisor(candle)
		if advice.DateTime.IsZero() {
			continue
		}
		if baseAdvice.DateTime.IsZero() {
			baseAdvice = advice
			lastAdvice = advice
			continue
		}
		if utils.IsNewFortsDateStarted(lastAdvice.DateTime, advice.DateTime) {
			var ds = DateSum{Date: DateTimeToDate(lastAdvice.DateTime), Sum: 1 + pnl/baseAdvice.Price}
			result = append(result, ds)
			pnl = 0
			baseAdvice = lastAdvice
		}
		if !skipPnl(lastAdvice.DateTime, advice.DateTime) {
			pnl += lastAdvice.Position*(advice.Price-lastAdvice.Price) -
				slippage*advice.Price*math.Abs(advice.Position-lastAdvice.Position)
		}
		lastAdvice = advice
	}

	if !lastAdvice.DateTime.IsZero() {
		var ds = DateSum{Date: DateTimeToDate(lastAdvice.DateTime), Sum: 1 + pnl/baseAdvice.Price}
		result = append(result, ds)
	}
	return result, nil
}

func MultiContractHprs(
	historyCandleService ICandleStorage,
	secCodes []string,
	advisorBuilder func() domain.Advisor,
	slippage float64,
	skipPnl func(time.Time, time.Time) bool,
	concurrency int,
) ([]DateSum, error) {
	if len(secCodes) == 1 {
		return SingleContractHprs(historyCandleService, secCodes[0], advisorBuilder, slippage, skipPnl)
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
				var hprs, err = SingleContractHprs(historyCandleService, securityCode, advisorBuilder, slippage, skipPnl)
				if err != nil {
					//return nil, err
					log.Println(err)
					continue
				}
				hprsByContracts[i] = hprs
			}
		}()
	}
	wg.Wait()
	var hprs = concatHprs(hprsByContracts)
	return hprs, nil
}

func AdvisorStatus(historyCandleService ICandleStorage,
	securityCode string,
	advisorBuilder func() domain.Advisor) error {

	var advisor = advisorBuilder()
	var advices []domain.Advice

	var err = historyCandleService.Walk(securityCode, func(candle domain.Candle) error {
		var advice = advisor(candle)
		if advice.DateTime.IsZero() {
			return nil
		}
		if len(advices) == 0 ||
			advices[len(advices)-1].Position != advice.Position {
			advices = append(advices, advice)
		}

		return nil
	})
	if err != nil {
		return err
	}

	const N = 10
	var skip = len(advices) - N
	if skip > 0 {
		advices = advices[skip:]
	}
	PrintAdvices(advices)
	return nil
}
