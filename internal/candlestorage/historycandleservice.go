package candlestorage

import (
	"advisordev/internal/core"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type CandleStorage struct {
	folderPath string
}

func NewCandleStorage(folderPath string) *CandleStorage {
	return &CandleStorage{folderPath: folderPath}
}

func (srv *CandleStorage) fileName(securityCode string) string {
	return filepath.Join(srv.folderPath, securityCode+".txt")
}

func (srv *CandleStorage) Walk(securityCode string,
	onCandle func(core.Candle) error) error {
	var path = srv.fileName(securityCode)
	var file, err = os.Open(path)
	if err != nil {
		return fmt.Errorf("CandleStorage.Walk %w", err)
	}
	defer file.Close()
	var reader = csv.NewReader(file)
	reader.Read()
	for {
		rec, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("CandleStorage.Walk %w", err)
		}
		candle, err := parseHistoryCandle(rec)
		if err != nil {
			return fmt.Errorf("CandleStorage.Walk %v %w", rec, err)
		}
		candle.SecurityCode = securityCode
		err = onCandle(candle)
		if err != nil {
			return err
		}
	}
}

func (srv *CandleStorage) Last(securityCode string) (core.Candle, error) {
	path := srv.fileName(securityCode)
	exists, err := isPathExists(path)
	if err != nil {
		return core.Candle{}, err
	}
	if !exists {
		return core.Candle{}, nil
	}

	var result core.Candle
	err = srv.Walk(securityCode, func(c core.Candle) error {
		result = c
		return nil
	})
	if err != nil {
		return core.Candle{}, err
	}
	return result, nil
}

func (srv *CandleStorage) Update(securityCode string, candles []core.Candle) error {
	f, err := os.OpenFile(srv.fileName(securityCode), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	csv := csv.NewWriter(f)
	//TODO header
	for _, c := range candles {
		record := []string{
			securityCode,
			"5",
			c.DateTime.Format("20060102"),
			strconv.Itoa(100 * (100*c.DateTime.Hour() + c.DateTime.Minute())),
			strconv.FormatFloat(c.OpenPrice, 'f', -1, 64),
			strconv.FormatFloat(c.HighPrice, 'f', -1, 64),
			strconv.FormatFloat(c.LowPrice, 'f', -1, 64),
			strconv.FormatFloat(c.ClosePrice, 'f', -1, 64),
			strconv.FormatFloat(c.Volume, 'f', -1, 64),
		}
		err := csv.Write(record)
		if err != nil {
			return err
		}
	}
	csv.Flush()
	return csv.Error()
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

func isPathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	// error other than not existing e.g. permission denied
	return false, err
}
