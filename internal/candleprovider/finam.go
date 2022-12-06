package candleprovider

import (
	"advisordev/internal/core"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	finamPeriodMinutes1 = 2
	finamPeriodMinutes5 = 3
	finamPeriodDay      = 8
)

func finamGet(client *http.Client,
	securityCode, periodCode int,
	beginDate, endDate time.Time) ([]core.Candle, error) {
	var url, err = historyCandlesFinamUrl(securityCode, periodCode, beginDate, endDate)
	if err != nil {
		return nil, err
	}
	//TODO log url?
	return getHistoryCandles(client, url)
}

func historyCandlesFinamUrl(securityCode int, periodCode int,
	beginDate, endDate time.Time) (string, error) {
	baseUrl, err := url.Parse("https://export.finam.ru/data.txt?d=d&market=14&f=data.txt&e=.txt&cn=data&dtf=1&tmf=1&MSOR=0&sep=1&sep2=1&datf=1&at=1")
	if err != nil {
		return "", err
	}

	params, err := url.ParseQuery(baseUrl.RawQuery)
	if err != nil {
		return "", err
	}

	params.Set("em", strconv.Itoa(securityCode))
	params.Set("df", strconv.Itoa(beginDate.Day()))
	params.Set("mf", strconv.Itoa(int(beginDate.Month())-1))
	params.Set("yf", strconv.Itoa(beginDate.Year()))
	params.Set("dt", strconv.Itoa(endDate.Day()))
	params.Set("mt", strconv.Itoa(int(endDate.Month())-1))
	params.Set("yt", strconv.Itoa(endDate.Year()))
	params.Set("p", strconv.Itoa(periodCode))

	baseUrl.RawQuery = params.Encode()
	return baseUrl.String(), nil
}
