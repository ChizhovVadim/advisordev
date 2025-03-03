package trader

import (
	"advisordev/internal/domain"
	"advisordev/internal/moex"
	"advisordev/internal/quik"
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"time"
)

type IMarketDataService interface {
	GetLastCandles(security domain.SecurityInfo, timeframe string) ([]domain.Candle, error)
	SubscribeCandles(security domain.SecurityInfo, timeframe string) error
}

type ISignalService interface {
	SubscribeMarketData() error
	OnMarketData(candle domain.Candle) domain.Advice
	ShowInfo() error
}

type IStrategyService interface {
	OnSignal(advice domain.Advice, outOrderRegistered *bool) error
	CheckPosition() bool
}

func Run(
	logger *slog.Logger,
	candleStorage domain.ICandleStorage,
	config TraderConfig,
) error {
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	var securityInformator = moex.NewFortsSecurityInformator()

	var (
		marketData        = make(chan domain.Candle)
		signals           []ISignalService
		strategies        []IStrategyService
		marketDataService IMarketDataService
	)

	for _, client := range config.Clients {
		if !(client.MarketData || hasActiveStrategies(client)) {
			continue
		}
		var logger = logger.With("client", client.Key)
		logger.Debug("Create trader")

		var trader domain.ITrader
		if client.Type == "quik" {
			var quik = quik.NewQuikConnector(logger, client.Port)
			defer quik.Close()
			trader = quik
			var err = quik.Init()
			if err != nil {
				return err
			}
			// эта горутина завершатся, тк defer quik.Close() закроет callback connection.
			go func() {
				// var candlesCallbackActive = &atomic.Bool{}
				quik.HandleCallbacks(ctx, marketData /*, candlesCallbackActive*/)
			}()
		} else if client.Type == "mock" {
			trader = NewMockTrader(logger)
		} else {
			// кроме quik можно поддержать API finam/alor/T.
			return fmt.Errorf("client type not supported %v", client.Type)
		}

		if client.MarketData {
			marketDataService = trader
		}

		var err = initStrategies(logger, trader, client.Portfolios, securityInformator, &strategies)
		if err != nil {
			return err
		}
	}

	var err = initSignals(logger, config.Signals, securityInformator, candleStorage, marketDataService, &signals)
	if err != nil {
		return err
	}

	var userCommands = make(chan string)
	go func() {
		defer close(userCommands)
		// команды из консоли не завершаются, поэтому не ждем завершения горутины
		// потом можно прикрутить, чтобы команды не только из консоли, но например из телеграм бота.
		readUserCommands(ctx, userCommands)
	}()

	return mainCycle(ctx, logger, strategies, signals, marketData, userCommands)
}

func hasActiveStrategies(client ClientConfig) bool {
	for _, p := range client.Portfolios {
		if len(p.Strategies) != 0 {
			return true
		}
	}
	return false
}

func initStrategies(
	logger *slog.Logger,
	trader domain.ITrader,
	portfolioConfigs []PortfolioConfig,
	securityInformator domain.ISecurityInformator,
	strategies *[]IStrategyService,
) error {
	connected, err := trader.IsConnected()
	if err != nil {
		return err
	}
	if !connected {
		return errors.New("trader is not connected")
	}

	for _, portfolioConfig := range portfolioConfigs {
		var logger = logger.With("portfolio", portfolioConfig.Portfolio)
		var portfolio = domain.PortfolioInfo{
			Firm:      portfolioConfig.Firm,
			Portfolio: portfolioConfig.Portfolio,
		}
		startAmount, err := trader.IncomingAmount(portfolio)
		if err != nil {
			return err
		}
		var availableAmount = startAmount
		if portfolioConfig.MaxAmount > 0 {
			availableAmount = math.Min(availableAmount, portfolioConfig.MaxAmount)
		}
		logger.Info("Init portfolio",
			"Amount", startAmount,
			"AvailableAmount", availableAmount)
		if availableAmount == 0 {
			logger.Warn("availableAmount zero")
			continue
		}

		for _, strategyConfig := range portfolioConfig.Strategies {
			security, err := securityInformator.GetSecurityInfo(strategyConfig.Security)
			if err != nil {
				return err
			}
			strategy, err := initStrategy(logger, strategyConfig, trader, portfolio, security, availableAmount)
			if err != nil {
				return err
			}
			*strategies = append(*strategies, strategy)
		}
	}
	return nil
}

func initSignals(
	logger *slog.Logger,
	signalConfigs []SignalConfig,
	securityInformator domain.ISecurityInformator,
	candleStorage domain.ICandleStorage,
	marketDataService IMarketDataService,
	signals *[]ISignalService,
) error {
	if marketDataService == nil {
		return errors.New("need at least one client with MarketData")
	}
	for _, signalConfig := range signalConfigs {
		var signal, err = initSignal(logger, signalConfig, securityInformator, candleStorage, marketDataService)
		if err != nil {
			return err
		}
		*signals = append(*signals, signal)
	}
	return nil
}

func readUserCommands(
	ctx context.Context,
	userCommands chan<- string,
) error {
	var scanner = bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var commandLine = scanner.Text()
		if commandLine == "quit" || commandLine == "exit" {
			return nil
		}
		if commandLine != "" {
			select {
			case <-ctx.Done():
			case userCommands <- commandLine:
			}
		}
	}
	return scanner.Err()
}

func mainCycle(
	ctx context.Context,
	logger *slog.Logger,
	strategies []IStrategyService,
	signals []ISignalService,
	marketData <-chan domain.Candle,
	userCommands <-chan string,
) error {

	var start = time.Now().Add(-10 * time.Minute)

	// здесь можно candlesCallbackActive.Store(true)
	// подписываемся как можно позже, перед чтением
	for _, signal := range signals {
		var err = signal.SubscribeMarketData()
		if err != nil {
			return err
		}
	}

	logger.Info("Init trader finished",
		"Strategy size", len(strategies),
		"Signal size", len(signals),
	)

	var checkPositionChan <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-checkPositionChan:
			checkPositionChan = nil
			for _, strategy := range strategies {
				strategy.CheckPosition()
			}

		case userCmd, ok := <-userCommands:
			if !ok {
				// exit user command
				return nil
			}
			if userCmd == "status" {
				for _, signal := range signals {
					signal.ShowInfo()
				}
				for _, strategy := range strategies {
					strategy.CheckPosition()
				}
			}

		case candle, ok := <-marketData:
			if !ok {
				marketData = nil
				continue
			}
			for _, signalService := range signals {
				var advice = signalService.OnMarketData(candle)
				if !advice.DateTime.IsZero() &&
					advice.DateTime.After(start) {

					// Может никто не подписан на сигнал, поэтому логируем
					logger.Debug("Advice changed", "Advice", advice)

					for _, strategy := range strategies {
						var orderRegistered bool
						var err = strategy.OnSignal(advice, &orderRegistered)
						if err != nil {
							logger.Error("OnSignal failed", "error", err)
						}
						if orderRegistered && checkPositionChan == nil {
							checkPositionChan = time.After(30 * time.Second)
						}
					}
				}
			}
		}
	}
}
