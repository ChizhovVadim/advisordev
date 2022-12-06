package candleprovider

import (
	"advisordev/internal/core"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	mfdPeriodMinutes5 = 2
	mfdPeriodDay      = 7
)

func mfdGet(client *http.Client,
	securityCode, periodCode int,
	beginDate, endDate time.Time) ([]core.Candle, error) {
	var url, err = historyCandlesMfdUrl(securityCode, periodCode, beginDate, endDate)
	if err != nil {
		return nil, err
	}
	//TODO log url?
	return getHistoryCandles(client, url)
}

func historyCandlesMfdUrl(securityCode, periodCode int,
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
	params.Set("Tickers", strconv.Itoa(securityCode))
	params.Set("Period", strconv.Itoa(periodCode))
	params.Set("StartDate", beginDate.Format(dateLayout))
	params.Set("EndDate", endDate.Format(dateLayout))

	baseUrl.RawQuery = params.Encode()
	return baseUrl.String(), nil
}
