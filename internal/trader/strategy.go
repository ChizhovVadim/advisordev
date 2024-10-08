package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/domain"
	"advisordev/internal/quik"
	"advisordev/internal/utils"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"
)

type Strategy struct {
	logger           *slog.Logger
	quikService      *quik.QuikService
	availableAmount  float64
	firm             string
	portfolio        string
	securityName     string
	securityCode     string
	advisor          domain.Advisor
	lastAdvice       domain.Advice
	strategyPosition int
	basePrice        float64
	secInfo          SecurityInfo
}

func initStrategy(
	logger *slog.Logger,
	quikService *quik.QuikService,
	availableAmount float64,
	firm string,
	portfolio string,
	strategyConfig advisors.StrategyConfig,
) (*Strategy, error) {
	if availableAmount == 0 {
		return nil, errors.New("availableAmount zero")
	}

	const interval = quik.CandleIntervalM5

	var advisor = advisors.Maindvisor(logger, strategyConfig)
	var securityName = strategyConfig.SecurityCode
	securityCode, err := utils.EncodeSecurity(securityName)
	if err != nil {
		return nil, err
	}
	lastQuikCandles, err := quikService.GetLastCandles(
		FuturesClassCode, securityCode, interval, 0)
	if err != nil {
		return nil, err
	}

	var lastCandles []domain.Candle
	for _, quikCandle := range lastQuikCandles {
		var candle = convertQuikCandle(securityName, quikCandle)
		//if !candle.DateTime.Before(skipBefore) {
		lastCandles = append(lastCandles, candle)
		//}
	}

	// последний бар за сегодня может быть не завершен
	if len(lastCandles) > 0 && isToday(lastCandles[len(lastCandles)-1].DateTime) {
		lastCandles = lastCandles[:len(lastCandles)-1]
	}

	if len(lastCandles) == 0 {
		logger.Warn("Ready candles empty")
	} else {
		logger.Info("Ready candles",
			"First", lastCandles[0],
			"Last", lastCandles[len(lastCandles)-1],
			"Size", len(lastCandles))
	}
	var initAdvice domain.Advice
	for _, candle := range lastCandles {
		var advice = advisor(candle)
		if !advice.DateTime.IsZero() {
			initAdvice = advice
		}
	}
	logger.Info("Init advice",
		"advice", initAdvice)

	pos, err := quikService.GetFuturesHolding(quik.GetFuturesHoldingRequest{
		FirmId:  firm,
		AccId:   portfolio,
		SecCode: securityCode,
	})
	if err != nil {
		return nil, err
	}
	var initPosition = int(pos.TotalNet)
	logger.Info("Init position",
		"Position", initPosition)

	err = quikService.SubscribeCandles(FuturesClassCode, securityCode, interval)
	if err != nil {
		return nil, err
	}
	return &Strategy{
		logger:           logger,
		quikService:      quikService,
		availableAmount:  availableAmount,
		firm:             firm,
		portfolio:        portfolio,
		securityName:     securityName,
		securityCode:     securityCode,
		advisor:          advisor,
		lastAdvice:       initAdvice,
		strategyPosition: initPosition,
		secInfo:          getSecurityInfo(securityName),
		basePrice:        0,
	}, nil
}

func (s *Strategy) OnNewCandle(
	newCandle quik.Candle,
) (bool, error) {
	if !(newCandle.Interval == quik.CandleIntervalM5 &&
		s.securityCode == newCandle.SecCode) {
		return false, nil
	}

	var myCandle = convertQuikCandle(s.securityName, newCandle)
	var advice = s.advisor(myCandle)
	if advice.DateTime.IsZero() {
		return false, nil
	}
	s.lastAdvice = advice
	if time.Since(advice.DateTime) >= 9*time.Minute {
		return false, nil
	}
	if s.basePrice == 0 {
		s.basePrice = myCandle.ClosePrice
		s.logger.Info("Init base price",
			"Advice", advice)
	}
	// размер лота
	var position = s.availableAmount / (s.basePrice * s.secInfo.Lever) * advice.Position
	var volume = int(position - float64(s.strategyPosition))
	if volume == 0 {
		return false, nil
	}
	s.logger.Info("New advice",
		"Advice", advice)
	err := s.CheckPosition()
	if err != nil {
		return false, err
	}
	err = s.registerOrder(volume, advice.Price, s.secInfo.PricePrecision)
	if err != nil {
		return false, fmt.Errorf("registerOrder failed %w", err)
	}
	s.strategyPosition += volume
	return true, nil
}

func (s *Strategy) CheckPosition() error {
	pos, err := s.quikService.GetFuturesHolding(quik.GetFuturesHoldingRequest{
		FirmId:  s.firm,
		AccId:   s.portfolio,
		SecCode: s.securityCode,
	})
	if err != nil {
		return err
	}
	var traderPosition = int(pos.TotalNet)
	if s.strategyPosition == traderPosition {
		s.logger.Info("Check position",
			"Position", s.strategyPosition,
			"Status", "+")
		return nil
	} else {
		s.logger.Warn("Check position",
			"StrategyPosition", s.strategyPosition,
			"TraderPosition", traderPosition,
			"Status", "!")
		return fmt.Errorf("StrategyPosition!=TraderPosition")
	}
}

func (s *Strategy) registerOrder(
	volume int,
	price float64,
	pricePrecision int,
) error {
	const Slippage = 0.001
	if volume > 0 {
		price = price * (1 + Slippage)
	} else {
		price = price * (1 - Slippage)
	}
	//TODO планка
	var sPrice = formatPrice(price, pricePrecision)
	s.logger.Info("Register order",
		"Price", sPrice,
		"Volume", volume)
	var trans = quik.Transaction{
		ACTION:    "NEW_ORDER",
		SECCODE:   s.securityCode,
		CLASSCODE: FuturesClassCode,
		ACCOUNT:   s.portfolio,
		PRICE:     sPrice,
	}
	if volume > 0 {
		trans.OPERATION = "B"
		trans.QUANTITY = strconv.Itoa(volume)
	} else {
		trans.OPERATION = "S"
		trans.QUANTITY = strconv.Itoa(-volume)
	}
	return s.quikService.SendTransaction(trans)
}

func formatPrice(price float64, pricePrecision int) string {
	return strconv.FormatFloat(price, 'f', pricePrecision, 64) //шаг цены
}
