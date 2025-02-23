package quik

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"advisordev/internal/domain"
	"advisordev/internal/moex"
)

type QuikConnector struct {
	logger       *slog.Logger
	port         int
	mainConn     net.Conn
	callbackConn net.Conn
	quikService  *QuikService
}

func NewQuikConnector(
	logger *slog.Logger,
	port int,
) *QuikConnector {
	return &QuikConnector{
		logger: logger,
		port:   port,
	}
}

func (c *QuikConnector) Close() error {
	var mainConnErr, callbackConnErr error
	if c.mainConn != nil {
		mainConnErr = c.mainConn.Close()
	}
	if c.callbackConn != nil {
		callbackConnErr = c.callbackConn.Close()
	}
	return errors.Join(mainConnErr, callbackConnErr)
}

func (c *QuikConnector) Init() error {
	mainConn, err := InitConnection(c.port)
	if err != nil {
		return err
	}
	c.mainConn = mainConn
	c.quikService = NewQuikService(mainConn)

	callbackConn, err := InitConnection(c.port + 1)
	if err != nil {
		return err
	}
	c.callbackConn = callbackConn

	return nil
}

func (c *QuikConnector) IsConnected() (bool, error) {
	return c.quikService.IsConnected()
}

// Входящие средства (Лимит открытых позиций для FORTS)
func (c *QuikConnector) IncomingAmount(
	portfolio domain.PortfolioInfo,
) (float64, error) {
	resp, err := c.quikService.GetPortfolioInfoEx(GetPortfolioInfoExRequest{
		FirmId:     portfolio.Firm,
		ClientCode: portfolio.Portfolio,
	})
	if err != nil {
		return 0, err
	}
	if !resp.Valid() {
		return 0, errors.New("portfolio not found")
	}
	return strconv.ParseFloat(resp.StartLimitOpenPos, 64)
}

func (c *QuikConnector) GetPosition(
	portfolio domain.PortfolioInfo,
	security domain.SecurityInfo,
) (float64, error) {
	if security.ClassCode == moex.FuturesClassCode {
		pos, err := c.quikService.GetFuturesHolding(GetFuturesHoldingRequest{
			FirmId:  portfolio.Firm,
			AccId:   portfolio.Portfolio,
			SecCode: security.Code,
		})
		if err != nil {
			return 0, err
		}
		return pos.TotalNet, nil
	} else {
		return 0, fmt.Errorf("not implemented")
	}
}

func (c *QuikConnector) RegisterOrder(
	order domain.Order,
) error {
	//TODO планка
	var sPrice = formatPrice(order.Security.PriceStep, order.Security.PricePrecision, order.Price)
	var trans = Transaction{
		ACTION:    "NEW_ORDER",
		SECCODE:   order.Security.Code,
		CLASSCODE: order.Security.ClassCode,
		ACCOUNT:   order.Portfolio.Portfolio,
		PRICE:     sPrice,
	}
	if order.Volume > 0 {
		trans.OPERATION = "B"
		trans.QUANTITY = strconv.Itoa(order.Volume)
	} else {
		trans.OPERATION = "S"
		trans.QUANTITY = strconv.Itoa(-order.Volume)
	}
	return c.quikService.SendTransaction(trans)
}

// TODO startTime time.Time
func (c *QuikConnector) GetLastCandles(
	security domain.SecurityInfo,
	timeframe string,
) ([]domain.Candle, error) {
	interval, err := quikTimeframe(timeframe)
	if err != nil {
		return nil, err
	}
	const count = 5_000 // Если не указывать размер, то может прийти слишком много баров и unmarshal большой json
	lastQuikCandles, err := c.quikService.GetLastCandles(security.ClassCode, security.Code, interval, count)
	if err != nil {
		return nil, err
	}

	var result []domain.Candle
	for _, quikCandle := range lastQuikCandles {
		var candle = convertQuikCandle(quikCandle)
		//if !candle.DateTime.Before(skipBefore) {
		result = append(result, candle)
		//}
	}

	// последний бар за сегодня может быть не завершен
	if len(result) > 0 && isToday(result[len(result)-1].DateTime) {
		result = result[:len(result)-1]
	}

	return result, nil
}

// TODO где-нибудь отписываться?
func (c *QuikConnector) SubscribeCandles(
	security domain.SecurityInfo,
	timeframe string,
) error {
	interval, err := quikTimeframe(timeframe)
	if err != nil {
		return err
	}
	isSubscribed, err := c.quikService.IsCandleSubscribed(security.ClassCode, security.Code, interval)
	if err != nil {
		return err
	}
	if !isSubscribed {
		err = c.quikService.SubscribeCandles(security.ClassCode, security.Code, interval)
		if err != nil {
			return err
		}
		c.logger.Debug("Subscribed",
			"security", security.Name,
			"timeframe", timeframe)
	}
	return nil
}

func (c *QuikConnector) LastPrice(security domain.SecurityInfo) (float64, error) {
	lastPriceParam, err := GetParamEx(c.quikService, security.ClassCode, security.Code, ParamNameLAST)
	if err != nil {
		return 0, err
	}
	return AsFloat64(lastPriceParam["param_value"])
}

func (c *QuikConnector) HandleCallbacks(
	ctx context.Context,
	candles chan<- domain.Candle,
) error {
	for cj, err := range QuikCallbacks(c.callbackConn) {
		if err != nil {
			return err
		}
		if cj.LuaError != "" {
			c.logger.Error("handleCallbacks",
				"LuaError", cj.LuaError)
			continue
		}
		if cj.Command == EventNameNewCandle {
			if cj.Data != nil && candles != nil {
				var newCandle Candle
				var err = json.Unmarshal(*cj.Data, &newCandle)
				if err != nil {
					return err
				}
				// TODO можно фильтровать слишком ранние бары
				select {
				case <-ctx.Done():
					return ctx.Err()
				case candles <- convertQuikCandle(newCandle):
				}
			}
			continue
		}
	}
	return nil
}
