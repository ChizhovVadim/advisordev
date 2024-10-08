package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/quik"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"strconv"
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

	gracefulShutdownCtx, gracefulShutdownCancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer gracefulShutdownCancel()

	mainConn, err := quik.InitConnection(client.Port)
	if err != nil {
		return err
	}
	defer mainConn.Close()

	var quikService = quik.NewQuikService(mainConn)

	callbackConn, err := quik.InitConnection(client.Port + 1)
	if err != nil {
		return err
	}
	defer callbackConn.Close()

	g, ctx := errgroup.WithContext(context.Background())

	var newCandles = make(chan quik.Candle, 16)

	g.Go(func() error {
		return handleCallbacks(ctx, logger, callbackConn, newCandles)
	})

	g.Go(func() error {
		defer callbackConn.Close() // сможет завершиться handleCallbacks!
		return runStrategies(ctx, gracefulShutdownCtx, logger, quikService, client, strategyConfigs, newCandles)
	})

	return g.Wait()
}

func handleCallbacks(
	ctx context.Context,
	logger *slog.Logger,
	r io.Reader,
	newCandles chan<- quik.Candle,
) error {
	for cj, err := range quik.QuikCallbacks(r) {
		if err != nil {
			return err
		}
		if cj.LuaError != "" {
			logger.Error("handleCallbacks",
				"LuaError", cj.LuaError)
			continue
		}
		if cj.Command == quik.EventNameNewCandle && cj.Data != nil {
			var newCandle quik.Candle
			var err = json.Unmarshal(*cj.Data, &newCandle)
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case newCandles <- newCandle:
			}
			continue
		}
	}
	return nil
}

func runStrategies(
	ctx context.Context,
	gracefulShutdownCtx context.Context,
	logger *slog.Logger,
	quikService *quik.QuikService,
	client Client,
	strategyConfigs []advisors.StrategyConfig,
	newCandles <-chan quik.Candle,
) error {
	logger.Info("Check connection")
	connected, err := quikService.IsConnected()
	if err != nil {
		return err
	}
	if !connected {
		return errors.New("quik is not connected")
	}

	logger = logger.With("portfolio", client.Portfolio)
	availableAmount, err := initPortfolio(logger, quikService, client)
	if err != nil {
		return err
	}

	// init strategies
	var strategies []IStrategy
	for _, strategyConfig := range strategyConfigs {
		var strategy IStrategy
		var strategyLogger = logger.With("security", strategyConfig.SecurityCode)
		if strategyConfig.Trader == "" {
			strategy, err = initStrategy(strategyLogger,
				quikService, availableAmount, client.Firm, client.Portfolio, strategyConfig)
		} else if strategyConfig.Trader == "quiet" {
			strategy, err = initQuietStrategy(strategyLogger, quikService, strategyConfig)
		} else {
			err = fmt.Errorf("bad trader %v", strategyConfig.Trader)
		}
		if err != nil {
			strategyLogger.Error("initStrategy failed",
				"err", err)
			continue
		}
		strategies = append(strategies, strategy)
	}

	var checkPositionChan <-chan time.Time

	// strategy cycle
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-gracefulShutdownCtx.Done():
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

func initPortfolio(
	logger *slog.Logger,
	quikService *quik.QuikService,
	client Client,
) (float64, error) {
	logger.Info("Init portfolio...")
	resp, err := quikService.GetPortfolioInfoEx(quik.GetPortfolioInfoExRequest{
		FirmId:     client.Firm,
		ClientCode: client.Portfolio,
	})
	if err != nil {
		return 0, err
	}
	if !resp.Valid() {
		return 0, errors.New("portfolio not found")
	}
	amount, err := strconv.ParseFloat(resp.StartLimitOpenPos, 64)
	if err != nil {
		return 0, err
	}
	var availableAmount float64
	if client.Amount > 0 {
		availableAmount = client.Amount
	} else {
		availableAmount = amount
	}
	if client.MaxAmount > 0 {
		availableAmount = math.Min(availableAmount, client.MaxAmount)
	}
	if 0 < client.Weight && client.Weight < 1 {
		availableAmount *= client.Weight
	}

	logger.Info("Init portfolio",
		"Amount", amount,
		"AvailableAmount", availableAmount)
	return availableAmount, nil
}
