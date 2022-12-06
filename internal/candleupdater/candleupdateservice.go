package candleupdater

import (
	"advisordev/internal/core"
	"fmt"
	"log"
	"time"
)

type ICandleStorage interface {
	Last(securityCode string) (core.Candle, error)
	Update(securityCode string, candles []core.Candle) error
}

type IHistoryCandleProvider interface {
	Load(securityCode string, beginDate, endDate time.Time) ([]core.Candle, error)
}

type CandleUpdateService struct {
	logger                *log.Logger
	candleStorage         ICandleStorage
	historyCandleProvider IHistoryCandleProvider
}

func NewCandleUpdateService(
	logger *log.Logger,
	candleStrorage ICandleStorage,
	historyCandleProvider IHistoryCandleProvider,
) *CandleUpdateService {
	return &CandleUpdateService{
		logger:                logger,
		candleStorage:         candleStrorage,
		historyCandleProvider: historyCandleProvider,
	}
}

func (srv *CandleUpdateService) Update(securityCode string) error {
	var expDate, ok = expirationDate(securityCode)
	if !ok {
		return fmt.Errorf("failed expiration date %v", securityCode)
	}

	var beginDate = expDate.AddDate(0, -4, 0)
	var endDate = time.Now()

	if beginDate.After(endDate) {
		return fmt.Errorf("new contract %v %v", securityCode, expDate)
	}

	var lastCandle, err = srv.candleStorage.Last(securityCode)
	if err != nil {
		return err
	}
	if !lastCandle.DateTime.IsZero() {
		if d := core.DateTimeToDate(lastCandle.DateTime); d.After(beginDate) {
			beginDate = d
		}
	}

	srv.logger.Printf("Downloading %v %v %v", securityCode, beginDate, endDate)

	candles, err := srv.historyCandleProvider.Load(securityCode, beginDate, endDate)
	if err != nil {
		//TODO можно вторую попытку через таймаут
		return err
	}

	if len(candles) > 0 {
		// Последний бар за сегодня может быть еще не завершен
		candles = candles[:len(candles)-1]
	}
	if !lastCandle.DateTime.IsZero() {
		candles = skipEarly(candles, lastCandle.DateTime)
	}
	if len(candles) == 0 {
		srv.logger.Printf("no new candles %v", securityCode)
		return nil
	}
	srv.logger.Printf("Downloaded %v %v", candles[0], candles[len(candles)-1])
	return srv.candleStorage.Update(securityCode, candles)
}
