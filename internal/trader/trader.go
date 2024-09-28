package trader

import (
	"advisordev/internal/advisors"
	"advisordev/internal/domain"
	"advisordev/internal/quik"
	"advisordev/internal/utils"
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

	"golang.org/x/sync/errgroup"
)

func Run(
	ctx context.Context,
	logger *slog.Logger,
	strategyConfigs []advisors.StrategyConfig,
	client Client,
	quietMode bool,
) error {
	logger.Info("trader::Run started.")
	defer logger.Info("trader::Run stopped.")

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

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return handleUserStopRequest(ctx, logger)
	})

	g.Go(func() error {
		<-ctx.Done()
		return callbackConn.Close()
	})

	var newCandles = make(chan quik.Candle, 16)

	g.Go(func() error {
		defer callbackConn.Close()
		return handleCallbacks(ctx, logger, callbackConn, newCandles)
	})

	g.Go(func() error {
		return runStrategies(ctx, logger, quikService, client, strategyConfigs, newCandles, quietMode)
	})

	return g.Wait()
}

func handleUserStopRequest(
	ctx context.Context,
	logger *slog.Logger,
) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c:
		logger.Info("Interrupt...")
		return errors.New("Interrupt")
	}
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

type StrategyEntry struct {
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

func runStrategies(
	ctx context.Context,
	logger *slog.Logger,
	quikService *quik.QuikService,
	client Client,
	strategyConfigs []advisors.StrategyConfig,
	newCandles <-chan quik.Candle,
	quietMode bool,
) error {
	logger.Info("Check connection")
	connected, err := quikService.IsConnected()
	if err != nil {
		return err
	}
	if !connected {
		return errors.New("quik is not connected")
	}

	var availableAmount float64
	if !quietMode {
		logger = logger.With("portfolio", client.Portfolio)
		availableAmount, err = initPortfolio(logger, quikService, client)
		if err != nil {
			return err
		}
	}

	// init strategies
	var strategies []StrategyEntry
	for _, strategyConfig := range strategyConfigs {
		var strategy, err = initStrategy(
			logger.With("security", strategyConfig.SecurityCode),
			quikService, client, strategyConfig)
		if err != nil {
			return err
		}
		strategies = append(strategies, strategy)
	}

	// strategy cycle
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case newCandle := <-newCandles:
			const interval = quik.CandleIntervalM5
			if newCandle.Interval == interval {
				for i := range strategies {
					var strategy = &strategies[i]
					if strategy.securityCode == newCandle.SecCode {
						var logger = logger.With("security", strategy.securityCode)
						if quietMode {
							var err = onNewCandle_QuietMode(logger, strategy, newCandle)
							if err != nil {
								logger.Error("onNewCandle_QuietMode",
									"err", err)
							}
						} else {
							var err = onNewCandle(logger, strategy, newCandle, availableAmount, quikService)
							if err != nil {
								logger.Error("onNewCandle",
									"err", err)
							}
						}
						break
					}
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
	if availableAmount == 0 {
		return 0, errors.New("availableAmount zero")
	}
	return availableAmount, nil
}

func initStrategy(
	logger *slog.Logger,
	quikService *quik.QuikService,
	client Client,
	strategyConfig advisors.StrategyConfig,
) (StrategyEntry, error) {
	const interval = quik.CandleIntervalM5

	var advisor = advisors.Maindvisor(logger, strategyConfig)
	var securityName = strategyConfig.SecurityCode
	securityCode, err := utils.EncodeSecurity(securityName)
	if err != nil {
		return StrategyEntry{}, err
	}
	lastQuikCandles, err := quikService.GetLastCandles(
		ClassCode, securityCode, interval, 0)
	if err != nil {
		return StrategyEntry{}, err
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
		FirmId:  client.Firm,
		AccId:   client.Portfolio,
		SecCode: securityCode,
	})
	if err != nil {
		return StrategyEntry{}, err
	}
	var initPosition = int(pos.TotalNet)
	logger.Info("Init position",
		"Position", initPosition)

	err = quikService.SubscribeCandles(ClassCode, securityCode, interval)
	if err != nil {
		return StrategyEntry{}, err
	}
	return StrategyEntry{
		firm:             client.Firm,
		portfolio:        client.Portfolio,
		securityName:     securityName,
		securityCode:     securityCode,
		advisor:          advisor,
		lastAdvice:       initAdvice,
		strategyPosition: initPosition,
		secInfo:          getSecurityInfo(securityName),
		basePrice:        0,
	}, nil
}

func onNewCandle_QuietMode(
	logger *slog.Logger,
	strategy *StrategyEntry,
	newCandle quik.Candle,
) error {
	var myCandle = convertQuikCandle(strategy.securityName, newCandle)
	var advice = strategy.advisor(myCandle)
	if advice.DateTime.IsZero() {
		return nil
	}
	logger.Info("New advice",
		"Advice", advice)
	return nil
}

func onNewCandle(
	logger *slog.Logger,
	strategy *StrategyEntry,
	newCandle quik.Candle,
	availableAmount float64,
	quikService *quik.QuikService,
) error {
	var myCandle = convertQuikCandle(strategy.securityName, newCandle)
	var advice = strategy.advisor(myCandle)
	if advice.DateTime.IsZero() {
		return nil
	}
	if strategy.basePrice == 0 {
		strategy.basePrice = myCandle.ClosePrice
		logger.Info("Init base price",
			"Advice", advice)
	}
	strategy.lastAdvice = advice
	// размер лота
	var position = availableAmount / (strategy.basePrice * strategy.secInfo.Lever) * advice.Position
	var volume = int(position - float64(strategy.strategyPosition))
	if volume == 0 {
		return nil
	}
	logger.Info("New advice",
		"Advice", advice)
	err := checkPosition(logger, strategy, quikService)
	if err != nil {
		return err
	}
	err = registerOrder(logger, quikService, strategy.portfolio, strategy.securityCode, volume, advice.Price, strategy.secInfo.PricePrecision)
	if err != nil {
		return fmt.Errorf("registerOrder failed %w", err)
	}
	strategy.strategyPosition += volume
	return nil
}

func checkPosition(
	logger *slog.Logger,
	strategy *StrategyEntry,
	quikService *quik.QuikService,
) error {
	pos, err := quikService.GetFuturesHolding(quik.GetFuturesHoldingRequest{
		FirmId:  strategy.firm,
		AccId:   strategy.portfolio,
		SecCode: strategy.securityCode,
	})
	if err != nil {
		return err
	}
	var traderPosition = int(pos.TotalNet)
	if strategy.strategyPosition == traderPosition {
		logger.Info("Check position",
			"Position", strategy.strategyPosition,
			"Status", "+")
		return nil
	} else {
		logger.Warn("Check position",
			"StrategyPosition", strategy.strategyPosition,
			"TraderPosition", traderPosition,
			"Status", "!")
		return fmt.Errorf("StrategyPosition!=TraderPosition")
	}
}

func registerOrder(
	logger *slog.Logger,
	quikService *quik.QuikService,
	portfolio, security string,
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
	logger.Info("Register order",
		"Price", sPrice,
		"Volume", volume)
	var trans = quik.Transaction{
		ACTION:    "NEW_ORDER",
		SECCODE:   security,
		CLASSCODE: ClassCode,
		ACCOUNT:   portfolio,
		PRICE:     sPrice,
	}
	if volume > 0 {
		trans.OPERATION = "B"
		trans.QUANTITY = strconv.Itoa(volume)
	} else {
		trans.OPERATION = "S"
		trans.QUANTITY = strconv.Itoa(-volume)
	}
	return quikService.SendTransaction(trans)
}

func formatPrice(price float64, pricePrecision int) string {
	return strconv.FormatFloat(price, 'f', pricePrecision, 64) //шаг цены
}
