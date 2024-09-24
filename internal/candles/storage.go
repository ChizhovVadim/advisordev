package candles

import (
	"advisordev/internal/domain"
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type CandleStorage struct {
	folderPath string
	loc        *time.Location
}

func NewCandleStorage(
	folderPath string,
	timeframe string,
	loc *time.Location,
) *CandleStorage {
	return &CandleStorage{
		folderPath: filepath.Join(folderPath, timeframe), //os.MkdirAll(folderPath, os.ModePerm)
		loc:        loc,
	}
}

func (srv *CandleStorage) fileName(securityCode string) string {
	return filepath.Join(srv.folderPath, securityCode+".txt")
}

// obsolete
func (srv *CandleStorage) Walk(securityCode string,
	onCandle func(domain.Candle) error) error {
	var path = srv.fileName(securityCode)
	var file, err = os.Open(path)
	if err != nil {
		return fmt.Errorf("CandleStorage.Walk %w", err)
	}
	defer file.Close()
	var reader = csv.NewReader(file)
	//reader.Comma = ';'
	reader.Read()
	for {
		rec, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("CandleStorage.Walk %w", err)
		}
		candle, err := parseCandleMetastock(rec, srv.loc)
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

func (srv *CandleStorage) Candles(
	securityCode string,
) iter.Seq2[domain.Candle, error] {
	return func(yield func(domain.Candle, error) bool) {
		var path = srv.fileName(securityCode)
		var file, err = os.Open(path)
		if err != nil {
			yield(domain.Candle{}, err)
			return
		}
		defer file.Close()
		var reader = csv.NewReader(file)
		//reader.Comma = ';'
		reader.Read()
		for {
			rec, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					return
				}
				yield(domain.Candle{}, err)
				return
			}
			candle, err := parseCandleMetastock(rec, srv.loc)
			if err != nil {
				yield(domain.Candle{}, err)
				return
			}
			candle.SecurityCode = securityCode
			if !yield(candle, nil) {
				return
			}
		}
	}
}

func (srv *CandleStorage) Last(securityCode string) (domain.Candle, error) {
	path := srv.fileName(securityCode)
	exists, err := isPathExists(path)
	if err != nil {
		return domain.Candle{}, err
	}
	if !exists {
		return domain.Candle{}, nil
	}

	var result domain.Candle
	err = srv.Walk(securityCode, func(c domain.Candle) error {
		result = c
		return nil
	})
	if err != nil {
		return domain.Candle{}, err
	}
	return result, nil
}

// Дописывает в конец файла
func (srv *CandleStorage) Update(securityCode string, candles []domain.Candle) error {
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
