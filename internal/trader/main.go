package trader

import (
	"advisordev/internal/domain"
	"advisordev/internal/moex"
	"advisordev/internal/quik"
	"context"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

// Разделяем generateSignals/executeSignals.
// Поэтому можно подменить signals и брать откуда-то с сервера (автоследование).
// connector используется в 2 горутинах, поэтому должен быть защищен mutex.
func Run(
	logger *slog.Logger,
	candleStorage domain.ICandleStorage,
	client Client,
	strategyConfigs []StrategyConfig,
) error {
	//var candleInterval = domain.CandleIntervalMinutes5
	var securityInformator domain.ISecurityInformator = moex.NewFortsSecurityInformator()

	var connector = quik.NewQuikConnector(logger, client.Port)
	defer connector.Close()
	var err = connector.Init()
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(context.Background())

	var candles = make(chan domain.Candle)
	var advices = make(chan domain.Advice)
	//var candlesCallbackActive = &atomic.Bool{}

	g.Go(func() error {
		defer close(candles)
		return connector.HandleCallbacks(ctx, candles /*, candlesCallbackActive*/)
	})

	g.Go(func() error {
		defer close(advices)
		return generateSignals(
			ctx,
			logger,
			strategyConfigs,
			securityInformator,
			candleStorage,
			connector,
			candles,
			advices)
	})

	g.Go(func() error {
		defer connector.Close()
		return executeSignals(
			ctx,
			logger,
			client,
			findSecuritiesForTrading(strategyConfigs),
			securityInformator,
			connector,
			advices,
		)
	})

	return g.Wait()
}

func findSecuritiesForTrading(strategyConfigs []StrategyConfig) []string {
	var securityNames []string
	for i := range strategyConfigs {
		var strategyConfig = &strategyConfigs[i]
		if strategyConfig.Trader != "quiet" {
			securityNames = append(securityNames, strategyConfig.SecurityCode)
		}
	}
	return securityNames
}
