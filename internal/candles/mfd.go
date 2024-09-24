package candles

import (
	"advisordev/internal/domain"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type MfdProvider struct {
	secCodes   map[string]string
	periodCode string
	client     *http.Client
}

func NewMfd(
	secCodes map[string]string,
	timeframe string,
	client *http.Client,
) (*MfdProvider, error) {
	var periodCode = mfdTimeFrame(timeframe)
	if periodCode == "" {
		return nil, fmt.Errorf("mfd timeFrameCode not found %v", timeframe)
	}
	return &MfdProvider{
		secCodes:   secCodes,
		periodCode: periodCode,
		client:     client,
	}, nil
}

func (srv *MfdProvider) Name() string {
	return "Mfd"
}

func (srv *MfdProvider) Load(securityName string, beginDate, endDate time.Time) ([]domain.Candle, error) {
	var secCode, ok = srv.secCodes[securityName]
	if !ok {
		return nil, fmt.Errorf("securityCode not found %v", securityName)
	}
	url, err := mfdUrl(secCode, srv.periodCode, beginDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("url failed %w", err)
	}
	res, err := getCandlesMatastock(srv.client, url)
	if err != nil {
		return nil, fmt.Errorf("getCandlesMatastock %v %w", url, err)
	}
	return res, nil
}

func mfdUrl(securityCode, periodCode string,
	beginDate, endDate time.Time) (string, error) {
	baseUrl, err := url.Parse("http://mfd.ru/export/handler.ashx/data.txt?TickerGroup=26&Alias=false&Period=2&timeframeValue=1&timeframeDatePart=day&SaveFormat=0&SaveMode=0&FileName=data.txt&FieldSeparator=%2C&DecimalSeparator=.&DateFormat=yyyyMMdd&TimeFormat=HHmmss&DateFormatCustom=&TimeFormatCustom=&AddHeader=true&RecordFormat=0&Fill=false")
	if err != nil {
		return "", err
	}

	params, err := url.ParseQuery(baseUrl.RawQuery)
	if err != nil {
		return "", err
	}

	const dateLayout = "02.01.2006"
	params.Set("Tickers", securityCode)
	params.Set("Period", periodCode)
	params.Set("StartDate", beginDate.Format(dateLayout))
	params.Set("EndDate", endDate.Format(dateLayout))

	baseUrl.RawQuery = params.Encode()
	return baseUrl.String(), nil
}

func mfdTimeFrame(tf string) string {
	if tf == TFMinutes5 {
		return "2"
	}
	if tf == TFHourly {
		return "6"
	}
	if tf == TFDaily {
		return "7"
	}
	return ""
}
