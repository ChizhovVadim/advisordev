package candles

import (
	"advisordev/internal/domain"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type FinamProvider struct {
	secCodes   map[string]string
	periodCode string
	client     *http.Client
}

func NewFinam(
	secCodes map[string]string,
	timeframe string,
	client *http.Client,
) (*FinamProvider, error) {
	var periodCode = finamTimeFrame(timeframe)
	if periodCode == "" {
		return nil, fmt.Errorf("finam timeFrameCode not found %v", timeframe)
	}
	return &FinamProvider{
		secCodes:   secCodes,
		periodCode: periodCode,
		client:     client,
	}, nil
}

func (srv *FinamProvider) Name() string {
	return "Finam"
}

func (srv *FinamProvider) Load(securityName string, beginDate, endDate time.Time) ([]domain.Candle, error) {
	var secCode, ok = srv.secCodes[securityName]
	if !ok {
		return nil, fmt.Errorf("securityCode not found %v", securityName)
	}
	url, err := finamUrl(secCode, srv.periodCode, beginDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("url failed %w", err)
	}
	res, err := getCandlesMatastock(srv.client, url)
	if err != nil {
		return nil, fmt.Errorf("getCandlesMatastock %v %w", url, err)
	}
	return res, nil
}

func finamUrl(securityCode, periodCode string,
	beginDate, endDate time.Time) (string, error) {
	baseUrl, err := url.Parse("https://export.finam.ru/data.txt?d=d&market=14&f=data.txt&e=.txt&cn=data&dtf=1&tmf=1&MSOR=0&sep=1&sep2=1&datf=1&at=1")
	if err != nil {
		return "", err
	}

	params, err := url.ParseQuery(baseUrl.RawQuery)
	if err != nil {
		return "", err
	}

	params.Set("em", securityCode)
	params.Set("df", strconv.Itoa(beginDate.Day()))
	params.Set("mf", strconv.Itoa(int(beginDate.Month())-1))
	params.Set("yf", strconv.Itoa(beginDate.Year()))
	params.Set("dt", strconv.Itoa(endDate.Day()))
	params.Set("mt", strconv.Itoa(int(endDate.Month())-1))
	params.Set("yt", strconv.Itoa(endDate.Year()))
	params.Set("p", periodCode)

	baseUrl.RawQuery = params.Encode()
	return baseUrl.String(), nil
}

func finamTimeFrame(tf string) string {
	if tf == domain.CandleIntervalMinutes5 {
		return "3"
	}
	if tf == domain.CandleIntervalHourly {
		return "7"
	}
	if tf == domain.CandleIntervalDaily {
		return "8"
	}
	return ""
}
