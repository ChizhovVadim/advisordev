package candleprovider

import (
	"advisordev/internal/core"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type SecurityCode struct {
	Code      string `xml:",attr"`
	FinamCode int    `xml:",attr"`
	MfdCode   int    `xml:",attr"`
}

type HistoryCandleProvider struct {
	logger        *log.Logger
	securityCodes []SecurityCode
	client        *http.Client
}

func NewHistoryCandleProvider(logger *log.Logger, securityCodes []SecurityCode) *HistoryCandleProvider {
	return &HistoryCandleProvider{
		logger:        logger,
		securityCodes: securityCodes,
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

func (srv *HistoryCandleProvider) Load(security string,
	beginDate, endDate time.Time) ([]core.Candle, error) {

	var securityCode, found = srv.findSecurity(security)
	if !found {
		return nil, fmt.Errorf("securityCode not found %v", securityCode)
	}
	var urls []string
	//TODO включать/откючать и порядок поставщиков через настройки
	if securityCode.FinamCode != 0 {
		var url, err = historyCandlesFinamUrl(securityCode.FinamCode, finamPeriodMinutes5, beginDate, endDate)
		if err == nil {
			urls = append(urls, url)
		}
	}
	if securityCode.MfdCode != 0 {
		var url, err = historyCandlesMfdUrl(securityCode.MfdCode, mfdPeriodMinutes5, beginDate, endDate)
		if err == nil {
			urls = append(urls, url)
		}
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("providers not found %v", securityCode)
	}
	var lastError error
	for _, url := range urls {
		srv.logger.Printf("getHistoryCandles from %v", url)
		result, err := getHistoryCandles(srv.client, url)
		if err != nil {
			srv.logger.Printf("getHistoryCandles failed %v", err)
			lastError = err
			continue
		}
		return result, nil
	}
	return nil, lastError
}

func (srv *HistoryCandleProvider) findSecurity(securityCode string) (SecurityCode, bool) {
	for i := range srv.securityCodes {
		if srv.securityCodes[i].Code == securityCode {
			return srv.securityCodes[i], true
		}
	}
	return SecurityCode{}, false
}

func getHistoryCandles(client *http.Client, url string) ([]core.Candle, error) {
	//resp, err := client.Get(url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %v", resp.Status)
	}

	var result []core.Candle

	csv := csv.NewReader(resp.Body)
	csv.Read() // skip fst line
	for {
		rec, err := csv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("getHistoryCandles %w", err)
		}
		c, err := parseHistoryCandle(rec)
		if err != nil {
			return nil, fmt.Errorf("getHistoryCandles %v %w", rec, err)
		}
		result = append(result, c)
	}
	if len(result) == 0 {
		return nil, core.ErrNoData
	}
	return result, nil
}

func parseHistoryCandle(record []string) (candle core.Candle, err error) {
	d, err := time.ParseInLocation("20060102", record[2], core.Moscow)
	if err != nil {
		return
	}
	t, err := strconv.Atoi(record[3])
	if err != nil {
		return
	}
	var hour = t / 10000
	var min = (t / 100) % 100
	d = d.Add(time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute)
	o, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		return
	}
	h, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		return
	}
	l, err := strconv.ParseFloat(record[6], 64)
	if err != nil {
		return
	}
	c, err := strconv.ParseFloat(record[7], 64)
	if err != nil {
		return
	}
	v, err := strconv.ParseFloat(record[8], 64)
	if err != nil {
		return
	}
	candle = core.Candle{DateTime: d, OpenPrice: o, HighPrice: h, LowPrice: l, ClosePrice: c, Volume: v}
	return
}
