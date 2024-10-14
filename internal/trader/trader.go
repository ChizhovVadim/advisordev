package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/domain"
	"advisordev/internal/quik"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"
)

func Run(
	logger *slog.Logger,
	strategyConfigs []advisors.StrategyConfig,
	client Client,
) error {
	logger.Info("trader::Run started.")
	defer logger.Info("trader::Run stopped.")

	var connector = quik.NewQuikConnector(logger, client.Port)
	defer connector.Close()

	var err = connector.Init()
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(context.Background())

	var newCandles = make(chan domain.Candle, 16)

	g.Go(func() error {
		return connector.HandleCallbacks(ctx, newCandles)
	})

	g.Go(func() error {
		defer connector.Close()
		return runStrategies(ctx, logger, connector, client, strategyConfigs, newCandles)
	})

	return g.Wait()
}

func runStrategies(
	ctx context.Context,
	logger *slog.Logger,
	connector domain.IConnector,
	client Client,
	strategyConfigs []advisors.StrategyConfig,
	newCandles <-chan domain.Candle,
) error {

	logger.Info("Check connection")
	connected, err := connector.IsConnected()
	if err != nil {
		return err
	}
	if !connected {
		return errors.New("quik is not connected")
	}

	logger.Info("Init portfolio")
	var portfolio = domain.PortfolioInfo{
		Firm:      client.Firm,
		Portfolio: client.Portfolio,
	}
	logger = logger.With("portfolio", portfolio.Portfolio)

	startAmount, err := connector.IncomingAmount(portfolio)
	if err != nil {
		return err
	}
	availableAmount := calcAvailableAmount(startAmount, client)
	logger.Info("Init portfolio",
		"Amount", startAmount,
		"AvailableAmount", availableAmount)

	// init strategies
	var strategies []IStrategy
	for _, strategyConfig := range strategyConfigs {
		var strategy IStrategy
		switch strategyConfig.Trader {
		case "quiet":
			strategy, err = initQuietStrategy(logger, connector, strategyConfig)
		case "":
			strategy, err = initStrategy(logger, connector, availableAmount, portfolio, strategyConfig)
		default:
			err = fmt.Errorf("bad trader %v", strategyConfig.Trader)
		}
		if err != nil {
			logger.Error("initStrategy failed",
				"strategyConfig", strategyConfig,
				"err", err)
			continue
		}
		strategies = append(strategies, strategy)
	}

	var checkPositionChan <-chan time.Time

	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt)

	// strategy cycle
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-interruptCh:
			logger.Info("graceful shutdown...")
			return nil
		case <-checkPositionChan:
			checkPositionChan = nil
			for i := range strategies {
				var strategy = strategies[i]
				strategy.CheckPosition()
			}
		case newCandle, ok := <-newCandles:
			if !ok {
				newCandles = nil
				continue
			}
			for i := range strategies {
				var strategy = strategies[i]
				orderRegistered, err := strategy.OnNewCandle(newCandle)
				if err != nil {
					logger.Error("handleNewCandle",
						"err", err)
					continue
				}
				if orderRegistered && checkPositionChan == nil {
					checkPositionChan = time.After(30 * time.Second)
				}
			}
		}
	}
}

func calcAvailableAmount(startAmount float64, client Client) float64 {
	var result float64
	if client.Amount > 0 {
		result = client.Amount
	} else {
		result = startAmount
	}
	if client.MaxAmount > 0 {
		result = math.Min(result, client.MaxAmount)
	}
	if 0 < client.Weight && client.Weight < 1 {
		result *= client.Weight
	}
	return result
}
