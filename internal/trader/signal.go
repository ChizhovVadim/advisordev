package trader

import (
	advisors "advisordev/internal/advisors_sample"
	"advisordev/internal/domain"
	"fmt"
	"log/slog"
)

type SignalService struct {
	logger            *slog.Logger
	marketDataService IMarketDataService
	security          domain.SecurityInfo
	candleInterval    string
	advisor           domain.Advisor
	lastAdvice        domain.Advice
}

func initSignal(
	logger *slog.Logger,
	config SignalConfig,
	securityInformator domain.ISecurityInformator,
	candleStorage domain.ICandleStorage,
	marketDataService IMarketDataService,
) (*SignalService, error) {
	logger = logger.With(
		"advisor", config.Advisor,
		"security", config.Security)

	security, err := securityInformator.GetSecurityInfo(config.Security)
	if err != nil {
		return nil, err
	}

	var candleInterval = domain.CandleIntervalMinutes5

	var advisor = advisors.MainAdvisor(config.Advisor, config.StdVolatility, logger)
	var initAdvice domain.Advice

	if candleStorage != nil {
		//TODO candleInterval
		for candle, err := range candleStorage.Candles(security.Name) {
			if err != nil {
				return nil, err
			}
			var advice = advisor(candle)
			if !advice.DateTime.IsZero() {
				initAdvice = advice
			}
		}
		logger.Debug("Init advice",
			"advice", initAdvice)
	}

	lastCandles, err := marketDataService.GetLastCandles(security, candleInterval)
	if err != nil {
		return nil, err
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

	advisor = applyStdDecorator(advisor, 0,
		config.Lever*config.Weight,
		config.MaxLever*config.Weight)

	// initAdvice без stdDecorator поэтому не сохраняем его в lastAdvice, чтобы не путаться
	return &SignalService{
		logger:            logger,
		marketDataService: marketDataService,
		security:          security,
		candleInterval:    candleInterval,
		advisor:           advisor,
	}, nil
}

func (signal *SignalService) SubscribeMarketData() error {
	return signal.marketDataService.SubscribeCandles(signal.security, signal.candleInterval)
}

func (signal *SignalService) OnMarketData(candle domain.Candle) domain.Advice {
	// советник следит только за своими барами
	//TODO проверять candleInterval
	if candle.SecurityCode != signal.security.Code {
		return domain.Advice{}
	}

	var advice = signal.advisor(candle)
	if !advice.DateTime.IsZero() {
		signal.lastAdvice = advice
	}
	return advice
}

func (signal *SignalService) ShowInfo() error {
	fmt.Println(signal.lastAdvice)
	return nil
}
