package update

import (
	"advisordev/internal/domain"
	"fmt"
	"iter"
	"log"
	"time"
)

type ICandleStorage interface {
	Candles(securityCode string) iter.Seq2[domain.Candle, error]
	Last(securityCode string) (domain.Candle, error)
	Update(securityCode string, candles []domain.Candle) error
}

func UpdateSignle(
	securityCode string,
	candleProvider ICandleProvider,
	candleStorage ICandleStorage,
	startDate func(securityCode string) time.Time,
	checkCandles func(l, r domain.Candle) error,
	maxDays int,
) error {
	var lastCandle, err = candleStorage.Last(securityCode)
	if err != nil {
		return err
	}
	var beginDate time.Time
	if lastCandle.DateTime.IsZero() {
		beginDate = startDate(securityCode)
	} else {
		beginDate = lastCandle.DateTime
	}
	var today = time.Now()
	var endDate = today

	// ограничение на кол-во скачиваемых данных за раз
	if maxDays != 0 {
		var limitDate = beginDate.AddDate(0, 0, maxDays)
		if limitDate.Before(endDate) {
			endDate = limitDate
		}
	}

	candles, err := candleProvider.Load(securityCode, beginDate, endDate)
	if err != nil {
		return err
	}
	if len(candles) == 0 {
		return fmt.Errorf("download empty %v", securityCode)
	}

	//Последний бар за сегодня может быть еще не завершен
	if fromOneDay(today, candles[len(candles)-1].DateTime) {
		candles = candles[:len(candles)-1]
	}

	if !lastCandle.DateTime.IsZero() {
		var startIndex = -1
		for i := range candles {
			if candles[i].DateTime.After(lastCandle.DateTime) {
				startIndex = i
				break
			}
		}
		if startIndex == -1 {
			candles = nil
		} else {
			candles = candles[startIndex:]
		}
	}

	if len(candles) == 0 {
		log.Println("No new candles",
			"securityCode", securityCode)
		return nil
	}

	if !lastCandle.DateTime.IsZero() && checkCandles != nil {
		var err = checkCandles(lastCandle, candles[0])
		if err != nil {
			return err
		}
	}

	log.Println("Downloaded",
		"provider", candleProvider.Name(),
		"securityCode", securityCode,
		"size", len(candles),
		"first", candles[0],
		"last", candles[len(candles)-1])

	//TODO отдельно?
	return candleStorage.Update(securityCode, candles)
}

func UpdateGroup(
	securityCodes []string,
	candleProviders []ICandleProvider,
	candleStorage ICandleStorage,
	startDate func(securityCode string) time.Time,
	checkCandles func(l, r domain.Candle) error,
	maxDays int,
) error {
	for _, candleProvider := range candleProviders {
		var providerName = candleProvider.Name()
		log.Println("UpdateGroup",
			"provider", providerName,
			"size", len(securityCodes))
		var secCodeFailed []string
		for _, secCode := range securityCodes {
			err := UpdateSignle(secCode, candleProvider, candleStorage, startDate, checkCandles, maxDays)
			if err != nil {
				log.Println("UpdateGroup",
					"provider", providerName,
					"secCode", secCode,
					"err", err)
				secCodeFailed = append(secCodeFailed, secCode)
			}
			time.Sleep(1 * time.Second)
		}
		if len(secCodeFailed) == 0 {
			return nil
		} else {
			log.Println("UpdateGroup failed",
				"provider", providerName,
				"size", len(secCodeFailed),
				"secCodeFailed", secCodeFailed,
			)
			securityCodes = secCodeFailed
		}
	}
	return fmt.Errorf("UpdateGroup failed")
}

func fromOneDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
