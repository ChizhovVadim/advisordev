package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/domain"
	"advisordev/internal/quik"
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
	logger = logger.
		With("security", strategyConfig.SecurityCode).
		With("advisor", strategyConfig.Name).
		With("strategy", "Strategy")

	if availableAmount == 0 {
		return nil, errors.New("availableAmount zero")
	}

	const interval = quik.CandleIntervalM5

	secInfo, err := getSecurityInfoHardCode(strategyConfig.SecurityCode)
	if err != nil {
		return nil, err
	}

	var advisor domain.Advisor
	var initAdvice domain.Advice
	err = initAdvisor(logger, quikService, strategyConfig, secInfo, &advisor, &initAdvice)
	if err != nil {
		return nil, err
	}

	pos, err := quikService.GetFuturesHolding(quik.GetFuturesHoldingRequest{
		FirmId:  firm,
		AccId:   portfolio,
		SecCode: secInfo.Code,
	})
	if err != nil {
		return nil, err
	}
	var initPosition = int(pos.TotalNet)
	logger.Info("Init position",
		"Position", initPosition)

	// TODO проверять не подписаны ли уже. где-нибудь отписываться?
	err = quikService.SubscribeCandles(secInfo.ClassCode, secInfo.Code, interval)
	if err != nil {
		return nil, err
	}

	return &Strategy{
		logger:           logger,
		quikService:      quikService,
		availableAmount:  availableAmount,
		firm:             firm,
		portfolio:        portfolio,
		secInfo:          secInfo,
		advisor:          advisor,
		lastAdvice:       initAdvice,
		strategyPosition: initPosition,
		basePrice:        0,
	}, nil
}

func (s *Strategy) OnNewCandle(
	newCandle quik.Candle,
) (bool, error) {
	if !(newCandle.Interval == quik.CandleIntervalM5 &&
		newCandle.SecCode == s.secInfo.Code) {
		return false, nil
	}

	var myCandle = convertQuikCandle(s.secInfo.Name, newCandle)
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
		SecCode: s.secInfo.Code,
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
		SECCODE:   s.secInfo.Code,
		CLASSCODE: s.secInfo.ClassCode,
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
