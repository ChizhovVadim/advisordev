package trader

import (
	advisors "advisordev/internal/advisors_sample"
	"advisordev/internal/domain"
	"context"
	"log/slog"
	"time"
)

func generateSignals(
	ctx context.Context,
	logger *slog.Logger,
	strategyConfigs []StrategyConfig,
	securityInformator domain.ISecurityInformator,
	candleStorage domain.ICandleStorage,
	connector domain.ITrader,
	candles <-chan domain.Candle,
	advices chan<- domain.Advice,
) error {
	logger = logger.With("name", "generateSignals")
	var start = time.Now().Add(-10 * time.Minute)
	var advisorList []Advisor
	for _, strategyConfig := range strategyConfigs {
		var advisor, err = initAdvisor(logger, strategyConfig, securityInformator, candleStorage, connector)
		if err != nil {
			return err
		}
		advisorList = append(advisorList, advisor)
	}
	// здесь можно candlesCallbackActive.Store(true)
	// подписываемся как можно позже, перед чтением
	for i := range advisorList {
		var err = advisorList[i].SubscribeCandles()
		if err != nil {
			return err
		}
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case candle, ok := <-candles:
			if !ok {
				return nil
			}
			for i := range advisorList {
				var advice = advisorList[i].OnNewCandle(candle)
				if advice.DateTime.IsZero() ||
					advice.DateTime.Before(start) {
					continue
				}
				logger.Debug("Advice changed",
					"advice", advice)
				if advices != nil {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case advices <- advice:
					}
				}
			}
		}
	}
}

type Advisor struct {
	logger    *slog.Logger
	connector domain.ITrader
	security  domain.SecurityInfo
	advisor   domain.Advisor
}

func initAdvisor(
	logger *slog.Logger,
	strategyConfig StrategyConfig,
	securityInformator domain.ISecurityInformator,
	candleStorage domain.ICandleStorage,
	connector domain.ITrader,
) (Advisor, error) {
	security, err := securityInformator.GetSecurityInfo(strategyConfig.SecurityCode)
	if err != nil {
		return Advisor{}, err
	}

	var advisor = advisors.MainAdvisor(strategyConfig.Name, strategyConfig.StdVolatility, logger)
	advisor = applyStdDecorator(
		advisor,
		strategyConfig.Direction,
		strategyConfig.Lever*strategyConfig.Weight,
		strategyConfig.MaxLever*strategyConfig.Weight)

	var initAdvice domain.Advice
	if candleStorage != nil {
		for candle, err := range candleStorage.Candles(strategyConfig.SecurityCode) {
			if err != nil {
				return Advisor{}, err
			}
			var advice = advisor(candle)
			if !advice.DateTime.IsZero() {
				initAdvice = advice
			}
		}
		logger.Debug("Init advice",
			"advice", initAdvice)
	}

	lastCandles, err := connector.GetLastCandles(security, domain.CandleIntervalMinutes5)
	if err != nil {
		return Advisor{}, err
	}

	if len(lastCandles) == 0 {
		logger.Warn("Ready candles empty")
	} else {
		logger.Debug("Ready candles",
			"First", lastCandles[0],
			"Last", lastCandles[len(lastCandles)-1],
			"Size", len(lastCandles))
	}

	for _, candle := range lastCandles {
		var advice = advisor(candle)
		if !advice.DateTime.IsZero() {
			initAdvice = advice
		}
	}
	logger.Info("Init advice",
		"advice", initAdvice)

	return Advisor{
		logger:    logger,
		connector: connector,
		security:  security,
		advisor:   advisor,
	}, nil
}

func (advisor *Advisor) SubscribeCandles() error {
	return advisor.connector.SubscribeCandles(advisor.security, domain.CandleIntervalMinutes5)
}

func (advisor *Advisor) OnNewCandle(candle domain.Candle) domain.Advice {
	// Можно еще проверять candleInterval
	if advisor.security.Code != candle.SecurityCode {
		return domain.Advice{}
	}
	return advisor.advisor(candle)
}
